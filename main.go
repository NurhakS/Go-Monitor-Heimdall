// Package main is the entry point for the uptime monitoring application
package main

import (
	"log"
	"strings"
	"time"
	"uptime-monitor/config"          // Configuration management
	"uptime-monitor/repository"      // Database operations
	"uptime-monitor/routes"          // HTTP routing
	"uptime-monitor/services"        // Application services
	scheduler "uptime-monitor/tasks" // Background monitoring tasks

	"github.com/gin-gonic/gin" // Web framework
)

// main is the entry point of the application
func main() {
	// Set log output to include more details
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Log start of application
	log.Println("🚀 Starting Uptime Monitor Application 🚀")

	// Initialize SQLite database and load configuration
	log.Println("Initializing configuration...")
	config.InitConfig()
	log.Println("Configuration initialized successfully")

	// Initialize repositories for different data types
	log.Println("Initializing repositories...")
	monitorRepo := repository.NewMonitorRepository(config.DB) // Handles website monitoring data
	logRepo := repository.NewLogRepository(config.DB)         // Handles monitoring logs
	smtpRepo := repository.NewSMTPRepository(config.DB)       // Handles email notification settings
	profileRepo := repository.NewProfileRepository(config.DB) // Handles user profiles
	log.Println("Repositories initialized successfully")

	// Initialize services
	log.Println("Initializing services...")
	services := services.NewServices(config.DB)
	log.Println("Services initialized successfully")

	// Start the background scheduler for monitoring websites
	log.Println("Starting background scheduler...")
	scheduler := scheduler.NewScheduler(services.Credentials, monitorRepo, logRepo)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Scheduler panic recovered: %v", r)
			}
		}()

		// Add more detailed logging
		log.Println("Starting scheduler in background goroutine...")

		// Start the scheduler
		scheduler.Start()

		// Keep the goroutine alive
		log.Println("Scheduler started, keeping goroutine alive...")
		for {
			// Log that the scheduler is still running every 5 minutes
			log.Println("Scheduler goroutine is still running...")
			time.Sleep(5 * time.Minute)
		}
	}()
	log.Println("Background scheduler started")

	// Initialize Gin web framework with default middleware
	log.Println("Initializing web server...")
	router := gin.Default()

	// Enable CORS (Cross-Origin Resource Sharing) for web access
	router.Use(func(c *gin.Context) {
		// Allow requests from any origin
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		// Allow specific HTTP methods
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// Allow specific headers
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Profile-ID")
		// Allow exposing specific headers to client
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		// Allow credentials (cookies, authorization headers)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		// Cache preflight requests for 24 hours
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		// Continue to next middleware/handler
		c.Next()
	})

	// Serve static files with cache control
	router.Use(func(c *gin.Context) {
		// Set proper cache control headers for static files
		if strings.HasPrefix(c.Request.URL.Path, "/static/") || strings.HasPrefix(c.Request.URL.Path, "/css/") {
			// Use a shorter cache time to allow for development changes
			c.Header("Cache-Control", "public, max-age=0")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
	})

	// Serve static files
	router.Static("/static", "./static")

	// Set up all application routes with their respective repositories
	log.Println("Setting up routes...")
	routes.SetupRoutes(router, monitorRepo, logRepo, smtpRepo, profileRepo, services.Credentials)
	log.Println("Routes set up successfully")

	// Start the HTTP server on port 8080
	log.Println("Starting HTTP server on port 8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("❌ Error starting server: ", err)
	}
}

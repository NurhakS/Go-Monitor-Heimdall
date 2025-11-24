package routes

import (
	"net/http"
	"uptime-monitor/controllers"
	"uptime-monitor/repository"
	"uptime-monitor/services"
	"uptime-monitor/types"

	"github.com/gin-gonic/gin"
)

// SetupRoutes initializes the API endpoints
func SetupRoutes(router *gin.Engine, monitorRepo *repository.MonitorRepository, logRepo *repository.LogRepository, smtpRepo *repository.SMTPRepository, profileRepo *repository.ProfileRepository, credentialsService *services.CredentialsService) {
	monitorController := controllers.NewMonitorController(monitorRepo)
	logController := controllers.NewLogController(logRepo)
	smtpController := controllers.NewSMTPController(smtpRepo)
	profileController := controllers.NewProfileController(profileRepo)
	credentialsController := controllers.NewCredentialsController(credentialsService)

	// Profile routes
	router.GET("/api/profiles", profileController.GetAllProfiles)
	router.POST("/api/profiles", profileController.CreateProfile)
	router.GET("/api/profiles/active", profileController.GetActiveProfile)
	router.GET("/api/profiles/:id", profileController.GetProfileByID)
	router.PUT("/api/profiles/:id", profileController.UpdateProfile)
	router.DELETE("/api/profiles/:id", profileController.DeleteProfile)
	router.POST("/api/profiles/:id/activate", profileController.SetActiveProfile)

	// Monitor routes
	router.GET("/api/monitors", monitorController.GetAllMonitors)
	router.POST("/api/monitors", monitorController.CreateMonitor)
	router.GET("/api/monitors/:id", monitorController.GetMonitor)
	router.PUT("/api/monitors/:id", monitorController.UpdateMonitor)
	router.DELETE("/api/monitors/:id", monitorController.DeleteMonitor)

	// Log routes
	router.POST("/logs", logController.CreateLog)
	router.GET("/logs/:monitor_id", logController.GetLogsByMonitor)

	// SMTP routes
	router.GET("/api/smtp_settings", smtpController.GetSMTPSettings)
	router.POST("/save_smtp", smtpController.UpdateSMTPSettings)
	router.DELETE("/api/smtp_settings/:id", smtpController.DeleteSMTPSettings)

	// Credentials routes
	router.GET("/api/credentials", credentialsController.GetCredentials)
	router.POST("/api/credentials", credentialsController.CreateCredential)
	router.GET("/api/credentials/:id", credentialsController.GetCredential)
	router.PUT("/api/credentials/:id", credentialsController.UpdateCredential)
	router.DELETE("/api/credentials/:id", credentialsController.DeleteCredential)

	// Notification methods routes
	router.GET("/api/notifications/methods", profileController.GetNotificationMethods)
	router.POST("/api/notifications/methods", profileController.CreateNotificationMethod)
	router.PUT("/api/notifications/methods/:id", profileController.UpdateNotificationMethod)
	router.DELETE("/api/notifications/methods/:id", profileController.DeleteNotificationMethod)

	// Static routes and pages
	router.LoadHTMLGlob("static/*.html")
	router.GET("/", func(c *gin.Context) {
		c.File("static/index.html")
	})
	router.GET("/notifications", func(c *gin.Context) {
		c.File("static/notifications.html")
	})
	router.GET("/profiles", func(c *gin.Context) {
		c.File("static/profiles.html")
	})
	router.GET("/credentials", func(c *gin.Context) {
		c.File("static/credentials.html")
	})
	router.Static("/css", "static/css")
}

func getMonitors(repo *repository.MonitorRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		monitors, err := repo.GetAllMonitors()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, monitors)
	}
}

func createMonitor(repo *repository.MonitorRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var monitor types.Monitor
		if err := c.ShouldBindJSON(&monitor); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := repo.CreateMonitor(&monitor); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, monitor)
	}
}

func updateMonitor(repo *repository.MonitorRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var monitor types.Monitor
		if err := c.ShouldBindJSON(&monitor); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		monitor.ID = id
		if err := repo.UpdateMonitor(&monitor); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, monitor)
	}
}

func deleteMonitor(repo *repository.MonitorRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := repo.DeleteMonitor(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Monitor deleted successfully"})
	}
}

package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"uptime-monitor/config"
	"uptime-monitor/repository"
	"uptime-monitor/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MonitorController struct {
	repo *repository.MonitorRepository
}

func NewMonitorController(repo *repository.MonitorRepository) *MonitorController {
	return &MonitorController{repo: repo}
}

// CreateMonitor creates a new monitor with SMTP details and sends a confirmation email
func (c *MonitorController) CreateMonitor(ctx *gin.Context) {
	var monitor types.Monitor
	if err := ctx.ShouldBindJSON(&monitor); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Debug logging
	fmt.Printf("Creating monitor: %+v\n", monitor)

	// Set default method if not provided
	if monitor.Method == "" {
		monitor.Method = "GET"
	}

	// Validate method
	validMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"HEAD":    true,
		"OPTIONS": true,
		"PATCH":   true,
	}

	if !validMethods[monitor.Method] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid HTTP method"})
		return
	}

	// Validate credential if provided
	if monitor.CredentialID != "" {
		// Check if credential exists and belongs to the active profile
		credRepo := repository.NewCredentialsRepository(config.DB)
		credential, err := credRepo.GetCredentialByID(monitor.CredentialID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credential"})
			return
		}

		// Ensure the credential belongs to the active profile
		activeProfile, err := repository.NewProfileRepository(config.DB).GetActiveProfile()
		if err != nil || credential.ProfileID != activeProfile.ID {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Credential does not belong to active profile"})
			return
		}
	}

	// Validate headers format if provided
	if monitor.Headers != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(monitor.Headers), &headers); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid headers format"})
			return
		}
		// Validate header values
		for key, value := range headers {
			if key == "" || value == "" {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid headers: empty key or value"})
				return
			}
		}
	}

	// Validate body format for POST/PUT requests
	if (monitor.Method == "POST" || monitor.Method == "PUT") && monitor.Body != "" {
		// Skip JSON validation for curl requests with form data
		if monitor.RequestType == "curl" && strings.Contains(monitor.Body, "--form") {
			// Form data for curl, no validation needed
			fmt.Printf("Skipping JSON validation for curl form data: %s\n", monitor.Body)
		} else {
			// For regular JSON bodies, validate the format
			var body map[string]interface{}
			if err := json.Unmarshal([]byte(monitor.Body), &body); err != nil {
				fmt.Printf("Invalid body format: %v\n", err)
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body format"})
				return
			}
		}
	}

	monitor.ID = uuid.New().String()
	monitor.CreatedAt = time.Now()
	monitor.UpdatedAt = time.Now()
	monitor.Status = "pending"
	monitor.IsActive = true
	monitor.FailureCount = 0

	if err := c.repo.CreateMonitor(&monitor); err != nil {
		fmt.Printf("Failed to create monitor: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create monitor"})
		return
	}

	fmt.Printf("Monitor created successfully: %s\n", monitor.ID)
	ctx.JSON(http.StatusCreated, monitor)
}

// GetMonitor retrieves a monitor by its ID
func (c *MonitorController) GetMonitor(ctx *gin.Context) {
	id := ctx.Param("id")
	monitor, err := c.repo.GetMonitorByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Monitor not found"})
		return
	}
	ctx.JSON(http.StatusOK, monitor)
}

// GetAllMonitors retrieves all monitors
func (c *MonitorController) GetAllMonitors(ctx *gin.Context) {
	monitors, err := c.repo.GetAllMonitors()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch monitors"})
		return
	}
	ctx.JSON(http.StatusOK, monitors)
}

func (c *MonitorController) DeleteMonitor(ctx *gin.Context) {
	id := ctx.Param("id")
	log.Printf("DELETE REQUEST: Deleting monitor with ID: %s", id)

	// Check if monitor exists first
	_, err := c.repo.GetMonitorByID(id)
	if err != nil {
		log.Printf("Monitor not found for deletion: %s - Error: %v", id, err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Monitor not found"})
		return
	}

	// Delete the monitor
	if err := c.repo.DeleteMonitor(id); err != nil {
		log.Printf("Failed to delete monitor %s: %v", id, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete monitor"})
		return
	}

	// Verify the monitor was actually deleted
	_, verifyErr := c.repo.GetMonitorByID(id)
	if verifyErr == nil {
		log.Printf("WARNING: Monitor %s still exists after deletion attempt!", id)
		// Try to delete again with a direct database query
		if db, ok := c.repo.GetDB().(*gorm.DB); ok {
			log.Printf("Attempting direct database deletion for monitor %s", id)
			result := db.Exec("DELETE FROM monitors WHERE id = ?", id)
			if result.Error != nil {
				log.Printf("Direct deletion error: %v", result.Error)
			} else {
				log.Printf("Direct deletion result: %d rows affected", result.RowsAffected)
			}
		}
	} else {
		log.Printf("Verified monitor %s was successfully deleted", id)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Monitor deleted successfully", "id": id})
}

// UpdateMonitor updates an existing monitor
func (c *MonitorController) UpdateMonitor(ctx *gin.Context) {
	id := ctx.Param("id")
	var monitor types.Monitor
	if err := ctx.ShouldBindJSON(&monitor); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate method
	validMethods := map[string]bool{
		"GET":    true,
		"POST":   true,
		"PUT":    true,
		"DELETE": true,
		"HEAD":   true,
	}

	if !validMethods[monitor.Method] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid HTTP method"})
		return
	}

	// Validate credential if provided
	if monitor.CredentialID != "" {
		// Check if credential exists and belongs to the active profile
		credRepo := repository.NewCredentialsRepository(config.DB)
		credential, err := credRepo.GetCredentialByID(monitor.CredentialID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credential"})
			return
		}

		// Ensure the credential belongs to the active profile
		activeProfile, err := repository.NewProfileRepository(config.DB).GetActiveProfile()
		if err != nil || credential.ProfileID != activeProfile.ID {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Credential does not belong to active profile"})
			return
		}
	}

	// Validate headers format if provided
	if monitor.Headers != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(monitor.Headers), &headers); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid headers format"})
			return
		}
		// Validate header values
		for key, value := range headers {
			if key == "" || value == "" {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid headers: empty key or value"})
				return
			}
		}
	}

	// Validate body format for POST/PUT requests
	if (monitor.Method == "POST" || monitor.Method == "PUT") && monitor.Body != "" {
		// Skip JSON validation for curl requests with form data
		if monitor.RequestType == "curl" && strings.Contains(monitor.Body, "--form") {
			// Form data for curl, no validation needed
		} else {
			// For regular JSON bodies, validate the format
			var body map[string]interface{}
			if err := json.Unmarshal([]byte(monitor.Body), &body); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body format"})
				return
			}
		}
	}

	// Set the ID from the URL parameter
	monitor.ID = id

	// Check if monitor exists and belongs to active profile
	existingMonitor, err := c.repo.GetMonitorByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Monitor not found"})
		return
	}

	// Preserve the profile ID from the existing monitor
	monitor.ProfileID = existingMonitor.ProfileID

	// Update the monitor
	if err := c.repo.UpdateMonitor(&monitor); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update monitor"})
		return
	}

	ctx.JSON(http.StatusOK, monitor)
}

package controllers

import (
	"log"
	"net/http"
	"time"
	"uptime-monitor/models"
	"uptime-monitor/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SMTPController struct {
	repo *repository.SMTPRepository
}

func NewSMTPController(repo *repository.SMTPRepository) *SMTPController {
	return &SMTPController{repo: repo}
}

// GetSMTPSettings retrieves all SMTP settings
func (c *SMTPController) GetSMTPSettings(ctx *gin.Context) {
	settings, err := c.repo.GetAllSMTPSettings()
	if err != nil {
		log.Printf("Error fetching SMTP settings: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch SMTP settings"})
		return
	}
	ctx.JSON(http.StatusOK, settings)
}

// UpdateSMTPSettings creates a new SMTP settings entry
func (c *SMTPController) UpdateSMTPSettings(ctx *gin.Context) {
	var settings models.SMTPSettings
	if err := ctx.ShouldBindJSON(&settings); err != nil {
		log.Printf("Error binding JSON: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if settings.SMTPHost == "" || settings.SMTPPort == "" || settings.SMTPEmail == "" ||
		settings.SMTPPassword == "" || settings.RecipientEmail == "" {
		log.Printf("Missing required fields in SMTP settings")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	// Generate a new ID and set creation time
	settings.ID = uuid.New().String()
	settings.CreatedAt = time.Now()

	log.Printf("Creating new SMTP settings: ID=%s, Host=%s, Email=%s",
		settings.ID, settings.SMTPHost, settings.SMTPEmail)

	if err := c.repo.UpdateSMTPSettings(&settings); err != nil {
		log.Printf("Error saving SMTP settings: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save SMTP settings"})
		return
	}

	log.Printf("Successfully created SMTP settings with ID: %s", settings.ID)
	ctx.JSON(http.StatusOK, gin.H{
		"message":  "SMTP settings saved successfully",
		"id":       settings.ID,
		"settings": settings,
	})
}

// DeleteSMTPSettings deletes a specific SMTP settings entry
func (c *SMTPController) DeleteSMTPSettings(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		log.Printf("Attempted to delete SMTP settings without ID")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	log.Printf("Deleting SMTP settings with ID: %s", id)
	if err := c.repo.DeleteSMTPSettingsByID(id); err != nil {
		log.Printf("Error deleting SMTP settings: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete SMTP settings"})
		return
	}

	log.Printf("Successfully deleted SMTP settings with ID: %s", id)
	ctx.JSON(http.StatusOK, gin.H{"message": "SMTP settings deleted successfully"})
}

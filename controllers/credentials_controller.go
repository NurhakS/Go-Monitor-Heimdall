package controllers

import (
	"log"
	"net/http"
	"time"
	"uptime-monitor/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CredentialsController struct {
	service *services.CredentialsService
}

func NewCredentialsController(service *services.CredentialsService) *CredentialsController {
	return &CredentialsController{service: service}
}

func (c *CredentialsController) GetCredentials(ctx *gin.Context) {
	// Comprehensive logging for debugging
	log.Printf("🔍 GetCredentials Request Received")
	log.Printf("  Full Request URL: %s", ctx.Request.URL.String())
	log.Printf("  Request Method: %s", ctx.Request.Method)

	// Log all incoming parameters for debugging
	log.Printf("  Query Parameters:")
	for key, values := range ctx.Request.URL.Query() {
		log.Printf("    %s: %v", key, values)
	}

	log.Printf("  Request Headers:")
	for key, values := range ctx.Request.Header {
		log.Printf("    %s: %v", key, values)
	}

	// Try to get profile ID from query parameter first
	profileID := ctx.Query("profileId")

	// If not in query parameter, try from header
	if profileID == "" {
		profileID = ctx.GetHeader("X-Profile-ID")
	}

	// Log the determined profile ID
	log.Printf("  Determined Profile ID: %s", profileID)

	// If still no profile ID, return error with detailed logging
	if profileID == "" {
		log.Println("❌ ERROR: No profile ID provided")
		log.Printf("  Available Query Params: %v", ctx.Request.URL.Query())
		log.Printf("  Available Headers: %v", ctx.Request.Header)

		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Profile ID is required",
			"details": gin.H{
				"query_params": ctx.Request.URL.Query(),
				"headers":      ctx.Request.Header,
			},
		})
		return
	}

	// Fetch credentials for the specific profile
	credentials, err := c.service.GetCredentials(profileID)
	if err != nil {
		log.Printf("❌ Error fetching credentials for profile %s: %v", profileID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch credentials",
			"details": gin.H{
				"profile_id":    profileID,
				"error_message": err.Error(),
			},
		})
		return
	}

	// Log the number of credentials found
	log.Printf("✅ Credentials found for profile %s: %d", profileID, len(credentials))

	// If no credentials found, return an empty array with logging
	if len(credentials) == 0 {
		log.Printf("⚠️ WARNING: No credentials found for Profile %s", profileID)
		ctx.JSON(http.StatusOK, []interface{}{})
		return
	}

	// Mask sensitive information before returning
	maskedCredentials := make([]gin.H, len(credentials))
	for i, cred := range credentials {
		maskedCredentials[i] = gin.H{
			"id":          cred.ID,
			"profile_id":  cred.ProfileID,
			"name":        cred.Name,
			"type":        cred.Type,
			"header_name": cred.HeaderName,
		}
	}

	// Return the masked credentials
	ctx.JSON(http.StatusOK, maskedCredentials)
}

func (c *CredentialsController) CreateCredential(ctx *gin.Context) {
	profileID := ctx.GetHeader("X-Profile-ID")
	if profileID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Profile ID is required"})
		return
	}

	var cred services.Credential
	if err := ctx.ShouldBindJSON(&cred); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credential data: " + err.Error()})
		return
	}

	cred.ID = uuid.New().String()
	cred.ProfileID = profileID
	cred.CreatedAt = time.Now()
	cred.UpdatedAt = time.Now()

	if err := c.service.CreateCredential(&cred); err != nil {
		log.Printf("Error creating credential: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create credential"})
		return
	}

	ctx.JSON(http.StatusCreated, cred)
}

func (c *CredentialsController) GetCredential(ctx *gin.Context) {
	profileID := ctx.GetHeader("X-Profile-ID")
	if profileID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Profile ID is required"})
		return
	}

	id := ctx.Param("id")
	cred, err := c.service.GetCredential(id)
	if err != nil {
		log.Printf("Error fetching credential %s: %v", id, err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Credential not found"})
		return
	}

	if cred.ProfileID != profileID {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	ctx.JSON(http.StatusOK, cred)
}

func (c *CredentialsController) UpdateCredential(ctx *gin.Context) {
	profileID := ctx.GetHeader("X-Profile-ID")
	if profileID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Profile ID is required"})
		return
	}

	id := ctx.Param("id")
	var cred services.Credential
	if err := ctx.ShouldBindJSON(&cred); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credential data: " + err.Error()})
		return
	}

	cred.ID = id
	cred.ProfileID = profileID
	cred.UpdatedAt = time.Now()

	if err := c.service.UpdateCredential(&cred); err != nil {
		log.Printf("Error updating credential %s: %v", id, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update credential"})
		return
	}

	ctx.JSON(http.StatusOK, cred)
}

func (c *CredentialsController) DeleteCredential(ctx *gin.Context) {
	profileID := ctx.GetHeader("X-Profile-ID")
	if profileID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Profile ID is required"})
		return
	}

	id := ctx.Param("id")
	if err := c.service.DeleteCredential(id, profileID); err != nil {
		log.Printf("Error deleting credential %s: %v", id, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete credential"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Credential deleted successfully"})
}

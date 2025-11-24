package controllers

import (
	"net/http"
	"uptime-monitor/repository"
	"uptime-monitor/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProfileController struct {
	repo *repository.ProfileRepository
}

func NewProfileController(repo *repository.ProfileRepository) *ProfileController {
	return &ProfileController{repo: repo}
}

func (c *ProfileController) CreateProfile(ctx *gin.Context) {
	var profile types.Profile
	if err := ctx.ShouldBindJSON(&profile); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.repo.CreateProfile(&profile); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile"})
		return
	}

	ctx.JSON(http.StatusCreated, profile)
}

func (c *ProfileController) CreateNotificationSettings(ctx *gin.Context) {
	var settings types.NotificationSettings
	if err := ctx.ShouldBindJSON(&settings); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get active profile
	activeProfile, err := c.repo.GetActiveProfile()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active profile"})
		return
	}

	// Set profile ID for the settings
	settings.ID = uuid.New().String()
	settings.ProfileID = activeProfile.ID

	if err := c.repo.CreateNotificationSettings(&settings); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification settings"})
		return
	}

	ctx.JSON(http.StatusCreated, settings)
}

func (c *ProfileController) GetAllProfiles(ctx *gin.Context) {
	profiles, err := c.repo.GetAllProfiles()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profiles"})
		return
	}
	ctx.JSON(http.StatusOK, profiles)
}

func (c *ProfileController) GetActiveProfile(ctx *gin.Context) {
	profile, err := c.repo.GetActiveProfile()
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No active profile found"})
		return
	}
	ctx.JSON(http.StatusOK, profile)
}

func (c *ProfileController) GetProfileByID(ctx *gin.Context) {
	id := ctx.Param("id")
	profile, err := c.repo.GetProfileByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	ctx.JSON(http.StatusOK, profile)
}

func (c *ProfileController) UpdateProfile(ctx *gin.Context) {
	id := ctx.Param("id")
	var profile types.Profile
	if err := ctx.ShouldBindJSON(&profile); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the ID from the URL parameter
	profile.ID = id

	if err := c.repo.UpdateProfile(&profile); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	ctx.JSON(http.StatusOK, profile)
}

func (c *ProfileController) DeleteProfile(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.repo.DeleteProfile(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete profile"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Profile deleted successfully", "id": id})
}

func (c *ProfileController) SetActiveProfile(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.repo.SetActiveProfile(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set active profile"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Profile activated successfully", "id": id})
}

func (c *ProfileController) GetNotificationMethods(ctx *gin.Context) {
	// Get active profile
	activeProfile, err := c.repo.GetActiveProfile()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active profile"})
		return
	}

	methods, err := c.repo.GetNotificationMethods(activeProfile.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notification methods"})
		return
	}

	ctx.JSON(http.StatusOK, methods)
}

func (c *ProfileController) CreateNotificationMethod(ctx *gin.Context) {
	var method types.NotificationMethod
	if err := ctx.ShouldBindJSON(&method); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get active profile
	activeProfile, err := c.repo.GetActiveProfile()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active profile"})
		return
	}

	// Set profile ID and generate ID
	method.ID = uuid.New().String()
	method.ProfileID = activeProfile.ID

	if err := c.repo.CreateNotificationMethod(&method); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification method"})
		return
	}

	ctx.JSON(http.StatusCreated, method)
}

func (c *ProfileController) UpdateNotificationMethod(ctx *gin.Context) {
	id := ctx.Param("id")
	var method types.NotificationMethod
	if err := ctx.ShouldBindJSON(&method); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the ID from the URL parameter
	method.ID = id

	if err := c.repo.UpdateNotificationMethod(&method); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification method"})
		return
	}

	ctx.JSON(http.StatusOK, method)
}

func (c *ProfileController) DeleteNotificationMethod(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.repo.DeleteNotificationMethod(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification method"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Notification method deleted successfully", "id": id})
}

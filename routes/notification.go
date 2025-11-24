package routes

import (
	"net/http"
	"uptime-monitor/repository"

	"github.com/gin-gonic/gin"
)

// SetupNotificationRoutes configures notification-related routes
func SetupNotificationRoutes(router *gin.Engine, notificationRepo *repository.NotificationRepository) {
	notifications := router.Group("/api/notifications")
	{
		// Get notification methods
		notifications.GET("/methods", func(c *gin.Context) {
			profileID := c.GetHeader("X-Profile-ID")
			if profileID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Profile ID is required"})
				return
			}

			methods, err := notificationRepo.GetNotificationMethods(profileID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notification methods: " + err.Error()})
				return
			}

			c.JSON(http.StatusOK, methods)
		})

		// Configure Slack notifications
		notifications.POST("/slack", func(c *gin.Context) {
			var req struct {
				WebhookURL string `json:"webhook_url" binding:"required"`
				Channel    string `json:"channel" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
				return
			}

			// Get active profile ID from context
			profileID := c.GetString("profile_id")
			if profileID == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - No profile ID found"})
				return
			}

			if err := notificationRepo.ConfigureSlack(profileID, req.WebhookURL, req.Channel); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to configure Slack: " + err.Error()})
				return
			}

			// Test the connection
			if err := notificationRepo.TestSlackConnection(profileID); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to test Slack connection: " + err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Slack notifications configured successfully"})
		})

		// Configure Teams notifications
		notifications.POST("/teams", func(c *gin.Context) {
			var req struct {
				WebhookURL string `json:"webhook_url" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
				return
			}

			// Get active profile ID from context
			profileID := c.GetString("profile_id")
			if profileID == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - No profile ID found"})
				return
			}

			if err := notificationRepo.ConfigureTeams(profileID, req.WebhookURL); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to configure Teams: " + err.Error()})
				return
			}

			// Test the connection
			if err := notificationRepo.TestTeamsConnection(profileID); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to test Teams connection: " + err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Teams notifications configured successfully"})
		})

		// Update notification method
		notifications.PATCH("/methods/:id", func(c *gin.Context) {
			id := c.Param("id")
			if id == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Method ID is required"})
				return
			}

			var req struct {
				Enabled bool `json:"enabled"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
				return
			}

			// Get the method
			method, err := notificationRepo.GetNotificationMethodByID(id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Notification method not found"})
				return
			}

			// Update the enabled status
			method.Enabled = req.Enabled
			if err := notificationRepo.UpdateNotificationMethod(method); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification method"})
				return
			}

			c.JSON(http.StatusOK, method)
		})

		// Delete notification method
		notifications.DELETE("/methods/:id", func(c *gin.Context) {
			id := c.Param("id")
			if id == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Method ID is required"})
				return
			}

			if err := notificationRepo.DeleteNotificationMethod(id); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification method"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Notification method deleted successfully"})
		})
	}
}

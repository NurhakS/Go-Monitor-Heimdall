package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"uptime-monitor/services"
	"uptime-monitor/types"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// ConfigureSlack sets up Slack notifications
func (r *NotificationRepository) ConfigureSlack(profileID string, webhookURL, channel string) error {
	if webhookURL == "" || channel == "" {
		return errors.New("webhook URL and channel are required")
	}

	config := map[string]string{
		"webhook_url": webhookURL,
		"channel":     channel,
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	method := &types.NotificationMethod{
		ID:        uuid.New().String(),
		ProfileID: profileID,
		Type:      "slack",
		Config:    configJSON,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Check if method already exists
	var existingMethod types.NotificationMethod
	if err := r.db.Where("type = ? AND profile_id = ?", "slack", profileID).First(&existingMethod).Error; err == nil {
		// Update existing method
		existingMethod.Config = configJSON
		existingMethod.UpdatedAt = time.Now()
		if err := r.db.Save(&existingMethod).Error; err != nil {
			return err
		}
		return nil
	}

	if err := r.db.Create(method).Error; err != nil {
		return err
	}
	return nil
}

// ConfigureTeams sets up Microsoft Teams notifications
func (r *NotificationRepository) ConfigureTeams(profileID string, webhookURL string) error {
	if webhookURL == "" {
		return errors.New("webhook URL is required")
	}

	config := map[string]string{
		"webhook_url": webhookURL,
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	method := &types.NotificationMethod{
		ID:        uuid.New().String(),
		ProfileID: profileID,
		Type:      "teams",
		Config:    configJSON,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Check if method already exists
	var existingMethod types.NotificationMethod
	if err := r.db.Where("type = ? AND profile_id = ?", "teams", profileID).First(&existingMethod).Error; err == nil {
		// Update existing method
		existingMethod.Config = configJSON
		existingMethod.UpdatedAt = time.Now()
		if err := r.db.Save(&existingMethod).Error; err != nil {
			return err
		}
		return nil
	}

	if err := r.db.Create(method).Error; err != nil {
		return err
	}
	return nil
}

// TestSlackConnection tests the Slack webhook
func (r *NotificationRepository) TestSlackConnection(profileID string) error {
	// Get Slack notification method
	methods, err := r.GetNotificationMethods(profileID)
	if err != nil {
		return fmt.Errorf("error getting notification methods: %v", err)
	}

	// Find Slack method
	var slackMethod *types.NotificationMethod
	for _, m := range methods {
		if m.Type == "slack" {
			slackMethod = &m
			break
		}
	}

	if slackMethod == nil {
		return errors.New("no Slack notification method found")
	}

	// Parse config
	var config map[string]interface{}
	if err := json.Unmarshal(slackMethod.Config, &config); err != nil {
		return fmt.Errorf("error parsing Slack config: %v", err)
	}

	// Create a test notification
	notificationService := services.NewNotificationService([]types.NotificationMethod{*slackMethod})
	return notificationService.SendSlack(&types.Monitor{
		Name: "Test Monitor",
		URL:  "http://test.com",
	}, "test", "This is a test notification from the uptime monitor", config)
}

// TestTeamsConnection tests the Teams webhook
func (r *NotificationRepository) TestTeamsConnection(profileID string) error {
	// Get Teams notification method
	methods, err := r.GetNotificationMethods(profileID)
	if err != nil {
		return fmt.Errorf("error getting notification methods: %v", err)
	}

	// Find Teams method
	var teamsMethod *types.NotificationMethod
	for _, m := range methods {
		if m.Type == "teams" {
			teamsMethod = &m
			break
		}
	}

	if teamsMethod == nil {
		return errors.New("no Teams notification method found")
	}

	// Parse config
	var config map[string]interface{}
	if err := json.Unmarshal(teamsMethod.Config, &config); err != nil {
		return fmt.Errorf("error parsing Teams config: %v", err)
	}

	// Create a test notification
	notificationService := services.NewNotificationService([]types.NotificationMethod{*teamsMethod})
	return notificationService.SendTeams(&types.Monitor{
		Name: "Test Monitor",
		URL:  "http://test.com",
	}, "test", "This is a test notification from the uptime monitor", config)
}

// GetNotificationMethodByID retrieves a notification method by its ID
func (r *NotificationRepository) GetNotificationMethodByID(id string) (*types.NotificationMethod, error) {
	if id == "" {
		return nil, errors.New("ID is required")
	}

	var method types.NotificationMethod
	if err := r.db.First(&method, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &method, nil
}

// GetNotificationMethods retrieves notification methods for a profile
func (r *NotificationRepository) GetNotificationMethods(profileID string) ([]types.NotificationMethod, error) {
	var methods []types.NotificationMethod
	err := r.db.Where("profile_id = ?", profileID).Find(&methods).Error
	if err != nil {
		return nil, err
	}
	return methods, nil
}

// UpdateNotificationMethod updates a notification method
func (r *NotificationRepository) UpdateNotificationMethod(method *types.NotificationMethod) error {
	if method == nil {
		return errors.New("method cannot be nil")
	}
	if method.ID == "" {
		return errors.New("method ID is required")
	}
	return r.db.Save(method).Error
}

// DeleteNotificationMethod removes a notification method
func (r *NotificationRepository) DeleteNotificationMethod(id string) error {
	if id == "" {
		return errors.New("ID is required")
	}
	return r.db.Delete(&types.NotificationMethod{}, "id = ?", id).Error
}

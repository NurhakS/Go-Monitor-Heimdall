package models

import "time"

// SMTPSettings represents email notification settings
type SMTPSettings struct {
	ID             string    `gorm:"type:varchar(36);primaryKey" json:"id"`       // Unique identifier
	ProfileID      string    `gorm:"type:varchar(36);not null" json:"profile_id"` // ID of the user profile
	SMTPHost       string    `gorm:"type:varchar(255);not null" json:"smtp_host"` // SMTP server hostname
	SMTPPort       string    `gorm:"type:varchar(10);not null" json:"smtp_port"`  // SMTP server port
	SMTPEmail      string    `gorm:"type:varchar(255)" json:"smtp_email"`         // SMTP username/email
	SMTPPassword   string    `gorm:"type:varchar(255)" json:"smtp_password"`      // SMTP password
	RecipientEmail string    `gorm:"type:varchar(255)" json:"recipient_email"`    // Email address to receive notifications
	IsActive       bool      `json:"is_active"`                                   // Whether email notifications are enabled
	CreatedAt      time.Time `gorm:"type:datetime" json:"created_at"`             // When the settings were created
	UpdatedAt      time.Time `gorm:"type:datetime" json:"updated_at"`             // When the settings were last updated
	Profile        Profile   `gorm:"foreignKey:ProfileID" json:"profile"`         // Associated user profile
}

// NotificationSettings represents all notification methods for a user
type NotificationSettings struct {
	ID        string `gorm:"type:varchar(36);primaryKey" json:"id"`       // Unique identifier
	ProfileID string `gorm:"type:varchar(36);not null" json:"profile_id"` // ID of the user profile

	// Email Settings
	EmailEnabled   bool   `json:"email_enabled"`                            // Whether email notifications are enabled
	SMTPHost       string `gorm:"type:varchar(255)" json:"smtp_host"`       // SMTP server hostname
	SMTPPort       string `gorm:"type:varchar(10)" json:"smtp_port"`        // SMTP server port
	SMTPEmail      string `gorm:"type:varchar(255)" json:"smtp_email"`      // SMTP username/email
	SMTPPassword   string `gorm:"type:varchar(255)" json:"smtp_password"`   // SMTP password
	RecipientEmail string `gorm:"type:varchar(255)" json:"recipient_email"` // Email address to receive notifications

	// Slack Settings
	SlackEnabled    bool   `json:"slack_enabled"`                              // Whether Slack notifications are enabled
	SlackWebhookURL string `gorm:"type:varchar(255)" json:"slack_webhook_url"` // Slack webhook URL
	SlackChannel    string `gorm:"type:varchar(255)" json:"slack_channel"`     // Slack channel to send notifications to

	// Microsoft Teams Settings
	TeamsEnabled    bool   `json:"teams_enabled"`                              // Whether Teams notifications are enabled
	TeamsWebhookURL string `gorm:"type:varchar(255)" json:"teams_webhook_url"` // Teams webhook URL

	IsActive  bool      `json:"is_active"`                           // Whether notifications are enabled
	CreatedAt time.Time `gorm:"type:datetime" json:"created_at"`     // When the settings were created
	UpdatedAt time.Time `gorm:"type:datetime" json:"updated_at"`     // When the settings were last updated
	Profile   Profile   `gorm:"foreignKey:ProfileID" json:"profile"` // Associated user profile
}

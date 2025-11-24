package types

import (
	"encoding/json"
	"time"
)

type NotificationMethod struct {
	ID        string          `json:"id"`
	ProfileID string          `json:"profile_id"`
	Type      string          `json:"type"`
	Enabled   bool            `json:"enabled"`
	Config    json.RawMessage `json:"config"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// ParseConfig helps parse the config into a specific type
func (nm *NotificationMethod) ParseConfig(v interface{}) error {
	if len(nm.Config) == 0 {
		return nil
	}
	return json.Unmarshal(nm.Config, v)
}

// EmailConfig represents email-specific notification configuration
type EmailConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	To       string `json:"to"`
}

// SlackConfig represents Slack-specific notification configuration
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
}

// TeamsConfig represents Microsoft Teams-specific notification configuration
type TeamsConfig struct {
	WebhookURL string `json:"webhook_url"`
}

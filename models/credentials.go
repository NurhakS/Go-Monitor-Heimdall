package models

import "time"

// Credentials represents stored authentication credentials
type Credentials struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	ProfileID   string    `json:"profile_id"`
	Name        string    `json:"name"`               // Display name for the credentials (e.g., "GitHub API Token")
	Type        string    `json:"type"`               // Type of credentials (e.g., "bearer", "basic", "api_key")
	Token       string    `json:"token"`              // The actual token/key value
	Username    string    `json:"username,omitempty"` // Username for basic auth
	Password    string    `json:"password,omitempty"` // Password for basic auth
	HeaderName  string    `json:"header_name"`        // Header name (e.g., "Authorization", "X-API-Key")
	HeaderValue string    `json:"header_value"`       // Header value template (e.g., "Bearer {token}")
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

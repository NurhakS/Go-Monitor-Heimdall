package models

import (
	"time"
)

type OAuth2Credentials struct {
	ID            string    `json:"id"`
	ProfileID     string    `json:"profile_id"`
	Name          string    `json:"name"`
	Provider      string    `json:"provider"` // e.g., "zoho", "google", etc.
	ClientID      string    `json:"client_id"`
	ClientSecret  string    `json:"client_secret"`
	RefreshToken  string    `json:"refresh_token"`
	AccessToken   string    `json:"access_token"`
	TokenType     string    `json:"token_type"`
	ExpiresAt     time.Time `json:"expires_at"`
	RedirectURI   string    `json:"redirect_uri"`
	LastRefreshed time.Time `json:"last_refreshed"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type OAuth2TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

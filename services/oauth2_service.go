package services

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"uptime-monitor/models"
)

type OAuth2Service struct {
	credentials *models.OAuth2Credentials
}

func NewOAuth2Service(credentials *models.OAuth2Credentials) *OAuth2Service {
	return &OAuth2Service{
		credentials: credentials,
	}
}

func (s *OAuth2Service) RefreshToken() error {
	// Build the curl command for token refresh
	curlCmd := exec.Command("curl",
		"--location",
		fmt.Sprintf("https://accounts.%s.com/oauth/v2/token", s.credentials.Provider),
		"--form", fmt.Sprintf("client_id=%s", s.credentials.ClientID),
		"--form", fmt.Sprintf("client_secret=%s", s.credentials.ClientSecret),
		"--form", fmt.Sprintf("redirect_uri=%s", s.credentials.RedirectURI),
		"--form", fmt.Sprintf("refresh_token=%s", s.credentials.RefreshToken),
		"--form", "grant_type=refresh_token",
	)

	// Execute the curl command
	output, err := curlCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to refresh token: %v, output: %s", err, string(output))
	}

	// Parse the response
	var tokenResponse models.OAuth2TokenResponse
	if err := json.Unmarshal(output, &tokenResponse); err != nil {
		return fmt.Errorf("failed to parse token response: %v", err)
	}

	// Update the credentials with the new token
	s.credentials.AccessToken = tokenResponse.AccessToken
	s.credentials.TokenType = tokenResponse.TokenType
	s.credentials.ExpiresAt = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
	s.credentials.LastRefreshed = time.Now()

	// If a new refresh token is provided, update it
	if tokenResponse.RefreshToken != "" {
		s.credentials.RefreshToken = tokenResponse.RefreshToken
	}

	return nil
}

func (s *OAuth2Service) GetAuthorizationHeader() string {
	// Check if token needs refresh (5 minutes buffer)
	if time.Now().Add(5 * time.Minute).After(s.credentials.ExpiresAt) {
		if err := s.RefreshToken(); err != nil {
			// Log the error but continue with current token
			fmt.Printf("Failed to refresh token: %v\n", err)
		}
	}

	return fmt.Sprintf("%s %s", strings.Title(s.credentials.TokenType), s.credentials.AccessToken)
}

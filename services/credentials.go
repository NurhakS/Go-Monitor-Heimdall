package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
)

type CredentialsService struct {
	db *gorm.DB
}

func NewCredentialsService(db *gorm.DB) *CredentialsService {
	return &CredentialsService{db: db}
}

type Credential struct {
	ID          string `json:"id"`
	ProfileID   string `json:"profile_id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Token       string `json:"token,omitempty"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	HeaderName  string `json:"header_name"`
	HeaderValue string `json:"header_value"`

	// OAuth2 Specific Fields
	ClientID     string    `json:"client_id,omitempty"`
	ClientSecret string    `json:"client_secret,omitempty"`
	RedirectURI  string    `json:"redirect_uri,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenExpiry  time.Time `json:"token_expiry,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *CredentialsService) RefreshOAuth2Token(cred *Credential) error {
	// Validate OAuth2 credential
	if cred.Type != "oauth2" ||
		cred.ClientID == "" ||
		cred.ClientSecret == "" ||
		cred.RefreshToken == "" {
		return fmt.Errorf("invalid OAuth2 credential")
	}

	// Prepare token refresh request
	data := url.Values{
		"client_id":     {cred.ClientID},
		"client_secret": {cred.ClientSecret},
		"refresh_token": {cred.RefreshToken},
		"grant_type":    {"refresh_token"},
	}

	// Make token refresh request
	resp, err := http.PostForm("https://accounts.zoho.com/oauth/v2/token", data)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %v", err)
	}
	defer resp.Body.Close()

	// Parse response
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token,omitempty"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %v", err)
	}

	// Prepare update
	updateData := map[string]interface{}{
		"token":        tokenResp.AccessToken,
		"updated_at":   time.Now(),
		"token_expiry": time.Now().Add(time.Second * time.Duration(tokenResp.ExpiresIn)),
	}

	// Update refresh token if provided
	if tokenResp.RefreshToken != "" {
		updateData["refresh_token"] = tokenResp.RefreshToken
	}

	// Update in database
	result := s.db.Model(&Credential{}).Where("id = ?", cred.ID).Updates(updateData)
	return result.Error
}

// GetCredential retrieves a credential by its ID
func (s *CredentialsService) GetCredential(credentialID string) (*Credential, error) {
	var cred Credential
	result := s.db.First(&cred, "id = ?", credentialID)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find credential: %v", result.Error)
	}
	return &cred, nil
}

// GetHeaderValue returns the appropriate header value based on credential type
func (s *CredentialsService) GetHeaderValue(cred *Credential) string {
	switch cred.Type {
	case "basic":
		return fmt.Sprintf("Basic %s", cred.Token)
	case "bearer":
		// Ensure the token doesn't already have "Bearer " prefix
		if strings.HasPrefix(cred.Token, "Bearer ") {
			return cred.Token
		}
		return fmt.Sprintf("Bearer %s", cred.Token)
	case "oauth2":
		// Ensure the token doesn't already have "Bearer " prefix
		if strings.HasPrefix(cred.Token, "Bearer ") {
			return cred.Token
		}
		return fmt.Sprintf("Bearer %s", cred.Token)
	case "api_key":
		return cred.Token
	default:
		return cred.HeaderValue
	}
}

// GetCredentials retrieves all credentials for a given profile ID
func (s *CredentialsService) GetCredentials(profileID string) ([]Credential, error) {
	// Comprehensive logging for credentials retrieval
	log.Printf("🔍 Retrieving Credentials for Profile ID: %s", profileID)

	// Validate profile ID
	if profileID == "" {
		log.Println("❌ ERROR: Empty profile ID provided")
		return nil, fmt.Errorf("empty profile ID")
	}

	// Fetch credentials for the specific profile with more detailed query
	var credentials []Credential
	result := s.db.Where("profile_id = ?", profileID).
		Select("id", "profile_id", "name", "type", "header_name", "header_value").
		Find(&credentials)

	// Log the database query details
	log.Printf("Database Query Details:")
	log.Printf("  Profile ID: %s", profileID)
	log.Printf("  Query Conditions: profile_id = %s", profileID)
	log.Printf("  Selected Fields: id, profile_id, name, type, header_name, header_value")

	// Check for database query errors
	if result.Error != nil {
		log.Printf("❌ Database Error retrieving credentials: %v", result.Error)
		return nil, result.Error
	}

	// Log the number of credentials found
	log.Printf("✅ Credentials Found for Profile %s: %d", profileID, len(credentials))

	// If no credentials found, log a warning
	if len(credentials) == 0 {
		log.Printf("⚠️ WARNING: No credentials found for Profile %s", profileID)
	}

	// Log the found credentials (without sensitive information)
	for _, cred := range credentials {
		log.Printf("  Credential Found:")
		log.Printf("    ID: %s", cred.ID)
		log.Printf("    Name: %s", cred.Name)
		log.Printf("    Type: %s", cred.Type)
		log.Printf("    Header Name: %s", cred.HeaderName)
	}

	return credentials, nil
}

// CreateCredential adds a new credential to the database
func (s *CredentialsService) CreateCredential(cred *Credential) error {
	return s.db.Create(cred).Error
}

// UpdateCredential updates an existing credential
func (s *CredentialsService) UpdateCredential(cred *Credential) error {
	return s.db.Model(&Credential{}).Where("id = ? AND profile_id = ?", cred.ID, cred.ProfileID).Updates(cred).Error
}

// DeleteCredential removes a credential by its ID and profile ID
func (s *CredentialsService) DeleteCredential(id, profileID string) error {
	return s.db.Where("id = ? AND profile_id = ?", id, profileID).Delete(&Credential{}).Error
}

package repository

import (
	"log"
	"uptime-monitor/models"

	"gorm.io/gorm"
)

type CredentialsRepository struct {
	db *gorm.DB
}

func NewCredentialsRepository(db *gorm.DB) *CredentialsRepository {
	return &CredentialsRepository{db: db}
}

func (r *CredentialsRepository) GetCredentialByID(id string) (*models.Credentials, error) {
	var credential models.Credentials
	result := r.db.First(&credential, "id = ?", id)
	if result.Error != nil {
		log.Printf("Error fetching credential %s: %v", id, result.Error)
		return nil, result.Error
	}
	return &credential, nil
}

func (r *CredentialsRepository) GetCredentialsByProfileID(profileID string) ([]models.Credentials, error) {
	var credentials []models.Credentials
	result := r.db.Where("profile_id = ?", profileID).Find(&credentials)
	if result.Error != nil {
		log.Printf("Error fetching credentials for profile %s: %v", profileID, result.Error)
		return nil, result.Error
	}
	return credentials, nil
}

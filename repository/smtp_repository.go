package repository

import (
	"log"
	"time"
	"uptime-monitor/models"

	"gorm.io/gorm"
)

type SMTPRepository struct {
	db *gorm.DB
}

func NewSMTPRepository(db *gorm.DB) *SMTPRepository {
	log.Println("Initializing SMTP repository...")
	return &SMTPRepository{db: db}
}

func (r *SMTPRepository) GetAllSMTPSettings() ([]models.SMTPSettings, error) {
	var settings []models.SMTPSettings

	// Get the active profile first
	var activeProfile models.Profile
	if err := r.db.Where("is_active = ?", true).First(&activeProfile).Error; err != nil {
		return nil, err
	}

	// Get SMTP settings for the active profile
	err := r.db.Where("profile_id = ?", activeProfile.ID).Find(&settings).Error
	return settings, err
}

func (r *SMTPRepository) GetSMTPSettings() (*models.SMTPSettings, error) {
	var settings models.SMTPSettings

	// Get the active profile first
	var activeProfile models.Profile
	if err := r.db.Where("is_active = ?", true).First(&activeProfile).Error; err != nil {
		return nil, err
	}

	// Get SMTP settings for the active profile
	err := r.db.Where("profile_id = ?", activeProfile.ID).First(&settings).Error
	return &settings, err
}

func (r *SMTPRepository) UpdateSMTPSettings(settings *models.SMTPSettings) error {
	// Get the active profile first
	var activeProfile models.Profile
	if err := r.db.Where("is_active = ?", true).First(&activeProfile).Error; err != nil {
		return err
	}

	// Set the profile ID for the new settings
	settings.ProfileID = activeProfile.ID
	settings.CreatedAt = time.Now()
	return r.db.Create(settings).Error
}

func (r *SMTPRepository) DeleteSMTPSettingsByID(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.SMTPSettings{}).Error
}

func (r *SMTPRepository) DeleteSMTPSettings() error {
	return r.db.Where("1=1").Delete(&models.SMTPSettings{}).Error
}

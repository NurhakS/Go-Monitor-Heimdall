package repository

import (
	"fmt"
	"log"
	"time"
	"uptime-monitor/types"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProfileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) CreateProfile(profile *types.Profile) error {
	profile.ID = uuid.New().String()
	profile.CreatedAt = time.Now()
	// Set all profiles to inactive first
	if err := r.db.Model(&types.Profile{}).Where("1=1").Update("is_active", false).Error; err != nil {
		return err
	}
	// Set the new profile as active
	profile.IsActive = true
	return r.db.Create(profile).Error
}

func (r *ProfileRepository) GetAllProfiles() ([]types.Profile, error) {
	var profiles []types.Profile
	err := r.db.Find(&profiles).Error
	return profiles, err
}

func (r *ProfileRepository) GetProfileByID(id string) (*types.Profile, error) {
	var profile types.Profile
	err := r.db.Where("id = ?", id).First(&profile).Error
	return &profile, err
}

func (r *ProfileRepository) UpdateProfile(profile *types.Profile) error {
	// Don't allow changing active status through update
	existingProfile := &types.Profile{}
	if err := r.db.Where("id = ?", profile.ID).First(existingProfile).Error; err != nil {
		return err
	}
	profile.IsActive = existingProfile.IsActive
	return r.db.Save(profile).Error
}

func (r *ProfileRepository) DeleteProfile(id string) error {
	// Don't allow deleting the active profile
	var profile types.Profile
	if err := r.db.Where("id = ?", id).First(&profile).Error; err != nil {
		return err
	}
	if profile.IsActive {
		return fmt.Errorf("cannot delete active profile")
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete associated monitors
		if err := tx.Where("profile_id = ?", id).Delete(&types.Monitor{}).Error; err != nil {
			return err
		}

		// Delete associated SMTP settings
		if err := tx.Where("profile_id = ?", id).Delete(&types.SMTPSettings{}).Error; err != nil {
			return err
		}

		// Delete the profile
		return tx.Where("id = ?", id).Delete(&types.Profile{}).Error
	})
}

func (r *ProfileRepository) GetActiveProfile() (*types.Profile, error) {
	var profile types.Profile
	result := r.db.Where("is_active = ?", true).First(&profile)
	if result.Error != nil {
		log.Printf("Error fetching active profile: %v", result.Error)
		return nil, result.Error
	}
	return &profile, nil
}

func (r *ProfileRepository) SetActiveProfile(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// First check if the profile exists
		var profile types.Profile
		if err := tx.Where("id = ?", id).First(&profile).Error; err != nil {
			return fmt.Errorf("profile not found")
		}

		// Deactivate all profiles
		if err := tx.Model(&types.Profile{}).Where("1=1").Update("is_active", false).Error; err != nil {
			return err
		}

		// Activate the selected profile
		return tx.Model(&types.Profile{}).Where("id = ?", id).Update("is_active", true).Error
	})
}

func (r *ProfileRepository) GetNotificationMethods(profileID string) ([]types.NotificationMethod, error) {
	var methods []types.NotificationMethod
	if err := r.db.Where("profile_id = ?", profileID).Find(&methods).Error; err != nil {
		return nil, err
	}
	return methods, nil
}

func (r *ProfileRepository) CreateNotificationMethod(method *types.NotificationMethod) error {
	return r.db.Create(method).Error
}

func (r *ProfileRepository) UpdateNotificationMethod(method *types.NotificationMethod) error {
	return r.db.Save(method).Error
}

func (r *ProfileRepository) DeleteNotificationMethod(id string) error {
	return r.db.Delete(&types.NotificationMethod{}, "id = ?", id).Error
}

func (r *ProfileRepository) CreateNotificationSettings(settings *types.NotificationSettings) error {
	return r.db.Create(settings).Error
}

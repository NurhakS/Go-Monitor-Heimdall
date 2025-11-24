package repository

import (
	"fmt"
	"log"
	"time"
	"uptime-monitor/types"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MonitorRepository struct {
	db *gorm.DB
}

func NewMonitorRepository(db *gorm.DB) *MonitorRepository {
	return &MonitorRepository{db: db}
}

func (r *MonitorRepository) CreateMonitor(monitor *types.Monitor) error {
	// Try to get the active profile first
	var activeProfile types.Profile
	if err := r.db.Where("is_active = ?", true).First(&activeProfile).Error; err != nil {
		// If no active profile exists, create a default one
		activeProfile = types.Profile{
			ID:          uuid.New().String(),
			Name:        "Default Profile",
			Description: "Automatically created default profile",
			IsActive:    true,
			CreatedAt:   time.Now(),
		}

		// Create the default profile
		if err := r.db.Create(&activeProfile).Error; err != nil {
			return fmt.Errorf("failed to create default profile: %v", err)
		}
	}

	// Set the profile ID for the new monitor
	monitor.ProfileID = activeProfile.ID
	return r.db.Create(monitor).Error
}

func (r *MonitorRepository) GetAllMonitors() ([]types.Monitor, error) {
	var monitors []types.Monitor

	// Get the active profile first
	var activeProfile types.Profile
	if err := r.db.Where("is_active = ?", true).First(&activeProfile).Error; err != nil {
		return nil, err
	}

	// Get monitors for the active profile
	// The monitors table doesn't have a deleted_at column, so we don't need to check for it
	log.Printf("Getting all monitors for profile %s", activeProfile.ID)
	err := r.db.Where("profile_id = ?", activeProfile.ID).Find(&monitors).Error

	log.Printf("Found %d monitors for profile %s", len(monitors), activeProfile.ID)
	return monitors, err
}

func (r *MonitorRepository) GetMonitorByID(id string) (*types.Monitor, error) {
	var monitor types.Monitor

	// Get the active profile first
	var activeProfile types.Profile
	if err := r.db.Where("is_active = ?", true).First(&activeProfile).Error; err != nil {
		return nil, err
	}

	// Get monitor for the active profile
	err := r.db.Where("id = ? AND profile_id = ?", id, activeProfile.ID).First(&monitor).Error
	return &monitor, err
}

func (r *MonitorRepository) UpdateMonitor(monitor *types.Monitor) error {
	log.Printf("UPDATING MONITOR: %s", monitor.Name)
	log.Printf("  New Status: %s", monitor.Status)
	log.Printf("  Response Code: %d", monitor.ResponseCode)
	log.Printf("  Response Time: %d", monitor.ResponseTime)

	// Use Save or Update method to ensure all fields are updated
	result := r.db.Save(monitor)
	if result.Error != nil {
		log.Printf("ERROR updating monitor: %v", result.Error)
		return result.Error
	}

	log.Printf("Monitor %s updated successfully", monitor.Name)
	return nil
}

func (r *MonitorRepository) DeleteMonitor(id string) error {
	// Get the active profile first
	var activeProfile types.Profile
	if err := r.db.Where("is_active = ?", true).First(&activeProfile).Error; err != nil {
		log.Printf("Error finding active profile for monitor deletion: %v", err)
		return err
	}

	// First try the standard GORM delete
	log.Printf("Deleting monitor with ID %s for profile %s", id, activeProfile.ID)
	result := r.db.Where("id = ? AND profile_id = ?", id, activeProfile.ID).Delete(&types.Monitor{})

	if result.Error != nil {
		log.Printf("Error deleting monitor %s: %v", id, result.Error)
		return result.Error
	}

	// If no rows were affected, try a direct SQL delete as a fallback
	if result.RowsAffected == 0 {
		log.Printf("No monitor found with ID %s for profile %s using GORM delete, trying direct SQL", id, activeProfile.ID)

		// Execute a direct SQL DELETE statement
		sqlResult := r.db.Exec("DELETE FROM monitors WHERE id = ? AND profile_id = ?", id, activeProfile.ID)

		if sqlResult.Error != nil {
			log.Printf("Error with direct SQL delete for monitor %s: %v", id, sqlResult.Error)
			return sqlResult.Error
		}

		if sqlResult.RowsAffected == 0 {
			log.Printf("No monitor found with ID %s for profile %s using direct SQL", id, activeProfile.ID)
			return fmt.Errorf("monitor not found")
		}

		log.Printf("Successfully deleted monitor %s using direct SQL (affected rows: %d)", id, sqlResult.RowsAffected)
	} else {
		log.Printf("Successfully deleted monitor %s using GORM (affected rows: %d)", id, result.RowsAffected)
	}

	// Double-check that the monitor is actually gone
	var count int64
	r.db.Model(&types.Monitor{}).Where("id = ?", id).Count(&count)

	if count > 0 {
		log.Printf("WARNING: Monitor %s still exists after deletion! Attempting final forced delete", id)
		// One final attempt with raw SQL and no conditions
		r.db.Exec("DELETE FROM monitors WHERE id = ?", id)

		// Check again
		r.db.Model(&types.Monitor{}).Where("id = ?", id).Count(&count)
		if count > 0 {
			log.Printf("CRITICAL: Monitor %s still exists after all deletion attempts!", id)
		} else {
			log.Printf("Final forced delete of monitor %s successful", id)
		}
	} else {
		log.Printf("Verified monitor %s is completely deleted from database", id)
	}

	return nil
}

// GetDB returns the underlying database connection
func (r *MonitorRepository) GetDB() interface{} {
	return r.db
}

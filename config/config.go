package config

import (
	"log"
	"time"
	"uptime-monitor/services"
	"uptime-monitor/types"

	"github.com/glebarez/sqlite" // This is a pure Go SQLite driver
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitConfig() {
	var err error
	DB, err = gorm.Open(sqlite.Open("monitor.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create tables with correct schema if they don't exist
	log.Println("Ensuring database schema is up to date...")
	err = DB.AutoMigrate(
		&types.Profile{},
		&types.Monitor{},
		&types.Log{},
		&types.SMTPSettings{},
		&types.NotificationSettings{},
		&types.NotificationMethod{},
		&services.Credential{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Create default profile if none exists
	var profileCount int64
	DB.Model(&types.Profile{}).Count(&profileCount)
	if profileCount == 0 {
		defaultProfile := types.Profile{
			ID:          uuid.New().String(),
			Name:        "Default Profile",
			Description: "Default monitoring profile",
			IsActive:    true,
			CreatedAt:   time.Now(),
		}
		if err := DB.Create(&defaultProfile).Error; err != nil {
			log.Printf("Failed to create default profile: %v", err)
		} else {
			log.Println("Created default profile")

			// Create default notification settings
			defaultSettings := types.NotificationSettings{
				ID:        uuid.New().String(),
				ProfileID: defaultProfile.ID,
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := DB.Create(&defaultSettings).Error; err != nil {
				log.Printf("Failed to create default notification settings: %v", err)
			} else {
				log.Println("Created default notification settings")
			}
		}
	}

	// Check if SMTP settings exist
	var smtpSettings types.SMTPSettings
	if DB.First(&smtpSettings).Error == gorm.ErrRecordNotFound {
		log.Println("No SMTP settings found. Please configure SMTP settings in the notifications page.")
	} else {
		log.Println("SMTP settings found in database")
	}

	// Check if notification settings exist
	var notificationSettings types.NotificationSettings
	if DB.First(&notificationSettings).Error == gorm.ErrRecordNotFound {
		log.Println("No notification settings found. Please configure notification settings in the notifications page.")
	} else {
		log.Println("Notification settings found in database")
	}

	log.Println("Database initialized successfully")
}

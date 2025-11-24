package services

import "gorm.io/gorm"

type Services struct {
	Credentials *CredentialsService
	// Add other services here as needed
}

func NewServices(db *gorm.DB) *Services {
	return &Services{
		Credentials: NewCredentialsService(db),
		// Initialize other services here
	}
}

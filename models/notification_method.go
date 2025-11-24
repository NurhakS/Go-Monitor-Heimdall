package models

import "time"

type NotificationMethod struct {
	ID        string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	ProfileID string    `gorm:"type:varchar(36);not null" json:"profile_id"`
	Type      string    `gorm:"type:varchar(50);not null" json:"type"` // email, slack, teams
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	Config    string    `gorm:"type:text" json:"config"` // JSON string containing method-specific configuration
	CreatedAt time.Time `gorm:"type:datetime" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:datetime" json:"updated_at"`
}

package models

import "time"

type Profile struct {
	ID          string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	IsActive    bool      `gorm:"default:false" json:"is_active"`
	CreatedAt   time.Time `gorm:"type:datetime" json:"created_at"`
}

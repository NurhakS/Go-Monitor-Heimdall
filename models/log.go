package models

import "time"

type Log struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	MonitorID string    `gorm:"not null" json:"monitor_id"`
	Status    string    `gorm:"not null" json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

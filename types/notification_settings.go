package types

import "time"

type NotificationSettings struct {
	ID        string    `json:"id"`
	ProfileID string    `json:"profile_id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

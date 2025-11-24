package types

import "time"

type Log struct {
	ID        string    `json:"id"`
	MonitorID string    `json:"monitor_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

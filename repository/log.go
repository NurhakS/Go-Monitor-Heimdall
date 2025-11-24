package repository

import (
	"uptime-monitor/types"

	"gorm.io/gorm"
)

type LogRepository struct {
	db *gorm.DB
}

func NewLogRepository(db *gorm.DB) *LogRepository {
	return &LogRepository{db: db}
}

func (r *LogRepository) CreateLog(log *types.Log) error {
	return r.db.Create(log).Error
}

func (r *LogRepository) GetLogsByMonitorID(monitorID string) ([]types.Log, error) {
	var logs []types.Log

	// Get the active profile first
	var activeProfile types.Profile
	if err := r.db.Where("is_active = ?", true).First(&activeProfile).Error; err != nil {
		return nil, err
	}

	// Verify the monitor belongs to the active profile
	var monitor types.Monitor
	if err := r.db.Where("id = ? AND profile_id = ?", monitorID, activeProfile.ID).First(&monitor).Error; err != nil {
		return nil, err
	}

	// Get logs for the monitor
	err := r.db.Where("monitor_id = ?", monitorID).Find(&logs).Error
	return logs, err
}

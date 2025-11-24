package services

import (
	"database/sql"
	"fmt"
	"time"
	"uptime-monitor/types"
)

type MonitorService struct {
	db          *sql.DB
	http        *HTTPService
	curl        *CurlService
	credentials *CredentialsService
	monitorRepo MonitorRepositoryInterface
}

type MonitorRepositoryInterface interface {
	CreateMonitor(monitor *types.Monitor) error
	GetAllMonitors() ([]types.Monitor, error)
	GetMonitorByID(id string) (*types.Monitor, error)
	UpdateMonitor(monitor *types.Monitor) error
	DeleteMonitor(id string) error
}

func NewMonitorService(
	db *sql.DB,
	http *HTTPService,
	curl *CurlService,
	credentials *CredentialsService,
	monitorRepo MonitorRepositoryInterface,
) *MonitorService {
	return &MonitorService{
		db:          db,
		http:        http,
		curl:        curl,
		credentials: credentials,
		monitorRepo: monitorRepo,
	}
}

func (s *MonitorService) CreateMonitor(monitor *types.Monitor) error {
	// Set default status to pending
	monitor.Status = "pending"

	// Set default check interval
	if monitor.CheckInterval < 10 {
		monitor.CheckInterval = 60
	}

	// Set default failure threshold
	if monitor.FailureThreshold < 1 {
		monitor.FailureThreshold = 1
	}

	// Ensure monitor is active
	monitor.IsActive = true

	// Set initial last checked time
	monitor.LastChecked = time.Now()

	// Use the monitorRepo to create the monitor
	return s.monitorRepo.CreateMonitor(monitor)
}

func (s *MonitorService) GetMonitor(id string) (*types.Monitor, error) {
	query := `
		SELECT id, profile_id, name, type, url, method, request_type,
			   headers, body, credential_id, check_interval, failure_threshold,
			   is_active, last_check_at, last_status, created_at, updated_at
		FROM monitors
		WHERE id = ?
	`

	var monitor types.Monitor
	err := s.db.QueryRow(query, id).Scan(
		&monitor.ID, &monitor.ProfileID, &monitor.Name, &monitor.Type,
		&monitor.URL, &monitor.Method, &monitor.RequestType,
		&monitor.Headers, &monitor.Body, &monitor.CredentialID,
		&monitor.CheckInterval, &monitor.FailureThreshold,
		&monitor.IsActive, &monitor.LastChecked, &monitor.Status,
		&monitor.CreatedAt, &monitor.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &monitor, nil
}

func (s *MonitorService) UpdateMonitor(monitor *types.Monitor) error {
	monitor.UpdatedAt = time.Now()

	query := `
		UPDATE monitors
		SET name = ?, type = ?, url = ?, method = ?, request_type = ?,
			headers = ?, body = ?, credential_id = ?, check_interval = ?,
			failure_threshold = ?, is_active = ?, updated_at = ?
		WHERE id = ? AND profile_id = ?
	`

	result, err := s.db.Exec(query,
		monitor.Name, monitor.Type, monitor.URL, monitor.Method,
		monitor.RequestType, monitor.Headers, monitor.Body,
		monitor.CredentialID, monitor.CheckInterval,
		monitor.FailureThreshold, monitor.IsActive, monitor.UpdatedAt,
		monitor.ID, monitor.ProfileID,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *MonitorService) DeleteMonitor(id, profileID string) error {
	query := "DELETE FROM monitors WHERE id = ? AND profile_id = ?"
	result, err := s.db.Exec(query, id, profileID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *MonitorService) GetMonitors(profileID string) ([]types.Monitor, error) {
	query := `
		SELECT id, profile_id, name, type, url, method, request_type,
			   headers, body, credential_id, check_interval, failure_threshold,
			   is_active, last_check_at, last_status, created_at, updated_at
		FROM monitors
		WHERE profile_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monitors []types.Monitor
	for rows.Next() {
		var monitor types.Monitor
		err := rows.Scan(
			&monitor.ID, &monitor.ProfileID, &monitor.Name, &monitor.Type,
			&monitor.URL, &monitor.Method, &monitor.RequestType,
			&monitor.Headers, &monitor.Body, &monitor.CredentialID,
			&monitor.CheckInterval, &monitor.FailureThreshold,
			&monitor.IsActive, &monitor.LastChecked, &monitor.Status,
			&monitor.CreatedAt, &monitor.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		monitors = append(monitors, monitor)
	}

	return monitors, nil
}

func (s *MonitorService) CheckMonitor(monitor *types.Monitor) (int, error) {
	var statusCode int
	var err error

	switch monitor.RequestType {
	case "curl":
		statusCode, message, err := s.curl.ExecuteCurlRequest(monitor)
		if err != nil {
			return 0, fmt.Errorf("curl request failed: %s", message)
		}
		return statusCode, nil
	default: // "http"
		resp, err := s.http.ExecuteRequest(monitor)
		if err != nil {
			return 0, fmt.Errorf("http request failed: %w", err)
		}
		defer resp.Body.Close()
		statusCode = resp.StatusCode
	}

	// Update last check time and status
	monitor.LastChecked = time.Now()
	monitor.Status = s.getStatusFromCode(statusCode)

	// Update monitor in database
	query := `
		UPDATE monitors
		SET last_check_at = ?, last_status = ?
		WHERE id = ?
	`
	_, err = s.db.Exec(query, monitor.LastChecked, monitor.Status, monitor.ID)
	if err != nil {
		return statusCode, fmt.Errorf("failed to update monitor status: %w", err)
	}

	return statusCode, nil
}

func (s *MonitorService) getStatusFromCode(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "up"
	case code == 401:
		return "unauthorized"
	default:
		return "down"
	}
}

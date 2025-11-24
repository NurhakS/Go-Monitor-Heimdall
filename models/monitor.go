package models

import (
	"encoding/json"
	"time"
)

// Monitor represents a website or service to monitor
type Monitor struct {
	// Basic Information
	ID               string             `json:"id" gorm:"primaryKey"`            // Unique identifier for the monitor
	Name             string             `json:"name"`                            // Display name of the monitor
	URL              string             `json:"url"`                             // URL or connection string to monitor
	Type             string             `json:"type"`                            // Type of monitor: "http", "mysql", "postgres", "mongodb", "redis"
	Method           string             `json:"method"`                          // HTTP method: GET, POST, etc.
	Headers          string             `json:"headers" gorm:"type:text"`        // JSON string of headers
	Body             string             `json:"body,omitempty" gorm:"type:text"` // JSON string of request body
	Interval         int                `json:"interval"`                        // Check interval in seconds
	Timeout          int                `json:"timeout"`                         // Timeout in seconds
	FailureThreshold int                `json:"failure_threshold"`               // Number of failures before alert
	Status           string             `json:"status"`                          // Current status: up, down, unauthorized
	FailureCount     int                `json:"failure_count"`                   // Current number of consecutive failures
	LastChecked      time.Time          `json:"last_checked"`                    // Last check timestamp
	ResponseTime     int64              `json:"response_time"`                   // Last response time in milliseconds
	ResponseCode     int                `json:"response_code"`                   // Last HTTP response code
	RequestType      string             `json:"request_type"`                    // Request type: "http" or "curl"
	CredentialID     string             `json:"credential_id,omitempty"`
	Credential       *OAuth2Credentials `json:"credential,omitempty" gorm:"-"`

	// Monitoring Settings
	CheckInterval  int  `json:"check_interval"`  // How often to check (in seconds)
	ExpectedStatus int  `json:"expected_status"` // Expected HTTP status code (e.g., 200)
	IsActive       bool `json:"is_active"`       // Whether monitoring is enabled

	// Status Information
	ProfileID string    `json:"profile_id"` // ID of the user profile that owns this monitor
	CreatedAt time.Time `json:"created_at"` // When the monitor was created
	UpdatedAt time.Time `json:"updated_at"` // When the monitor was last updated

	// Database-specific fields
	DBHost          string `json:"db_host"`           // Database host address
	DBPort          string `json:"db_port"`           // Database port number
	DBName          string `json:"db_name"`           // Database name
	DBUsername      string `json:"db_username"`       // Database username
	DBPassword      string `json:"db_password"`       // Database password
	DBQuery         string `json:"db_query"`          // SQL query or MongoDB command to execute
	DBExpectedValue string `json:"db_expected_value"` // Expected result from query
}

// UnmarshalJSON custom unmarshaler for Monitor
func (m *Monitor) UnmarshalJSON(data []byte) error {
	type Alias Monitor
	aux := &struct {
		Headers json.RawMessage `json:"headers"`
		Body    json.RawMessage `json:"body"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle Headers
	if aux.Headers != nil {
		var headers map[string]string
		if err := json.Unmarshal(aux.Headers, &headers); err == nil {
			headersJSON, _ := json.Marshal(headers)
			m.Headers = string(headersJSON)
		}
	}

	// Handle Body
	if aux.Body != nil {
		var body map[string]interface{}
		if err := json.Unmarshal(aux.Body, &body); err == nil {
			bodyJSON, _ := json.Marshal(body)
			m.Body = string(bodyJSON)
		}
	}

	return nil
}

// GetHeadersMap returns the headers as a map
func (m *Monitor) GetHeadersMap() map[string]string {
	if m.Headers == "" {
		return nil
	}
	var headers map[string]string
	_ = json.Unmarshal([]byte(m.Headers), &headers)
	return headers
}

// GetBodyMap returns the body as a map
func (m *Monitor) GetBodyMap() map[string]interface{} {
	if m.Body == "" {
		return nil
	}
	var body map[string]interface{}
	_ = json.Unmarshal([]byte(m.Body), &body)
	return body
}

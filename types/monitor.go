package types

import (
	"encoding/json"
	"fmt"
	"time"
)

type Monitor struct {
	ID               string    `json:"id"`
	ProfileID        string    `json:"profile_id"`
	Name             string    `json:"name"`
	Type             string    `json:"type"`
	URL              string    `json:"url"`
	Method           string    `json:"method"`
	RequestType      string    `json:"request_type"`
	Headers          string    `json:"headers"`
	Body             string    `json:"body"`
	CredentialID     string    `json:"credential_id"`
	CheckInterval    int       `json:"check_interval"`
	FailureThreshold int       `json:"failure_threshold"`
	FailureCount     int       `json:"failure_count"`
	Timeout          int       `json:"timeout"`
	IsActive         bool      `json:"is_active"`
	Status           string    `json:"status"`
	ResponseCode     int       `json:"response_code"`
	ResponseTime     int64     `json:"response_time"`
	LastChecked      time.Time `json:"last_checked"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Database-specific fields
	DBHost          string `json:"db_host,omitempty"`
	DBPort          string `json:"db_port,omitempty"`
	DBName          string `json:"db_name,omitempty"`
	DBUsername      string `json:"db_username,omitempty"`
	DBPassword      string `json:"db_password,omitempty"`
	DBQuery         string `json:"db_query,omitempty"`
	DBExpectedValue string `json:"db_expected_value,omitempty"`
}

// GetHeadersMap parses the Headers string into a map
func (m *Monitor) GetHeadersMap() map[string]string {
	if m.Headers == "" {
		return nil
	}

	headers := make(map[string]string)
	err := json.Unmarshal([]byte(m.Headers), &headers)
	if err != nil {
		fmt.Printf("Error parsing headers: %v\n", err)
		return nil
	}

	return headers
}

// GetBodyMap parses the Body string into a map
func (m *Monitor) GetBodyMap() map[string]interface{} {
	if m.Body == "" {
		return nil
	}

	// If the body is form data (for curl), return nil
	if m.RequestType == "curl" && m.Body != "" && m.Body[0] != '{' {
		return nil
	}

	body := make(map[string]interface{})
	err := json.Unmarshal([]byte(m.Body), &body)
	if err != nil {
		fmt.Printf("Error parsing body: %v\n", err)
		return nil
	}

	return body
}

// IsCurlRequest returns true if this monitor uses curl
func (m *Monitor) IsCurlRequest() bool {
	return m.RequestType == "curl"
}

// IsFormData returns true if the body is form data for curl
func (m *Monitor) IsFormData() bool {
	return m.IsCurlRequest() && m.Body != "" && m.Body[0] != '{' && m.Body[0] != '['
}

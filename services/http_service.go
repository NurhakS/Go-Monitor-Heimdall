package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"uptime-monitor/types"
)

type HTTPService struct {
	client      *http.Client
	credentials *CredentialsService
}

func NewHTTPService(credentials *CredentialsService) *HTTPService {
	return &HTTPService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		credentials: credentials,
	}
}

func (s *HTTPService) ExecuteRequest(monitor *types.Monitor) (*http.Response, error) {
	// Create request
	var body io.Reader
	if bodyMap := monitor.GetBodyMap(); bodyMap != nil {
		jsonBody, err := json.Marshal(bodyMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		body = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(monitor.Method, monitor.URL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers if specified
	if headers := monitor.GetHeadersMap(); headers != nil {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	// Add credential headers if specified
	if monitor.CredentialID != "" {
		cred, err := s.credentials.GetCredential(monitor.CredentialID)
		if err != nil {
			return nil, fmt.Errorf("failed to get credential: %v", err)
		}

		// Get the processed header value with placeholders replaced
		headerValue := s.credentials.GetHeaderValue(cred)
		req.Header.Set(cred.HeaderName, headerValue)
	}

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

func (s *HTTPService) CheckEndpoint(monitor *types.Monitor) (*types.Log, error) {
	startTime := time.Now()

	// Create request
	req, err := http.NewRequest(monitor.Method, monitor.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Add headers
	if headers := monitor.GetHeadersMap(); headers != nil {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}

	// Add body for POST/PUT requests
	if (monitor.Method == "POST" || monitor.Method == "PUT") && monitor.GetBodyMap() != nil {
		jsonBody, err := json.Marshal(monitor.GetBodyMap())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	}

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return &types.Log{
			ID:        monitor.ID,
			MonitorID: monitor.ID,
			Status:    "error",
			Message:   fmt.Sprintf("Request failed: %v", err),
			CreatedAt: time.Now(),
		}, nil
	}
	defer resp.Body.Close()

	// Calculate response time
	responseTime := time.Since(startTime).Milliseconds()

	// Create log entry
	log := &types.Log{
		ID:        monitor.ID,
		MonitorID: monitor.ID,
		Status:    "success",
		Message:   fmt.Sprintf("Response status: %d", resp.StatusCode),
		CreatedAt: time.Now(),
	}

	// Update monitor status
	monitor.Status = "up"
	monitor.ResponseTime = responseTime
	monitor.LastChecked = time.Now()

	return log, nil
}

package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
	"uptime-monitor/types"
)

// CurlService handles HTTP requests using Go's native HTTP client or the curl command
type CurlService struct {
	credentials *CredentialsService
	client      *http.Client
}

func NewCurlService(credentials *CredentialsService) *CurlService {
	return &CurlService{
		credentials: credentials,
		client: &http.Client{
			Timeout: 10 * time.Second, // Default timeout
		},
	}
}

// ExecuteCurlRequest executes a request using either Go's HTTP client or the curl command
func (s *CurlService) ExecuteCurlRequest(monitor *types.Monitor) (int, string, error) {
	// If the request type is "curl", use the actual curl command
	if strings.ToLower(monitor.RequestType) == "curl" {
		return s.executeCurlCommand(monitor)
	}

	// Otherwise, use the Go HTTP client
	return s.executeHttpRequest(monitor)
}

// executeHttpRequest executes a request using Go's HTTP client
func (s *CurlService) executeHttpRequest(monitor *types.Monitor) (int, string, error) {
	// Create request
	var body io.Reader
	if monitor.Body != "" {
		body = strings.NewReader(monitor.Body)
	}

	req, err := http.NewRequest(monitor.Method, monitor.URL, body)
	if err != nil {
		return 0, fmt.Sprintf("Request creation failed: %v", err), err
	}

	// Add headers
	if headers := parseHeaders(monitor.Headers); headers != nil {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}

	// Add credential headers if specified
	if monitor.CredentialID != "" {
		cred, err := s.credentials.GetCredential(monitor.CredentialID)
		if err != nil {
			return 0, fmt.Sprintf("Credential retrieval failed: %v", err), err
		}

		headerValue := s.credentials.GetHeaderValue(cred)
		req.Header.Add(cred.HeaderName, headerValue)
	}

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return 0, fmt.Sprintf("Connection failed to %s: %v", monitor.URL, err), err
	}
	defer resp.Body.Close()

	// Debug logging
	fmt.Printf("Monitor URL: %s\n", monitor.URL)
	fmt.Printf("Monitor Method: %s\n", monitor.Method)
	fmt.Printf("Request Headers: %v\n", req.Header)
	fmt.Printf("Response Status Code: %d\n", resp.StatusCode)

	// Generate appropriate message based on status code
	statusCode := resp.StatusCode
	var message string
	switch {
	case statusCode >= 200 && statusCode < 300:
		message = fmt.Sprintf("%s returned status code %d (Success)", monitor.URL, statusCode)
	case statusCode == 401:
		message = fmt.Sprintf("%s returned status code 401 (Unauthorized)", monitor.URL)
	case statusCode == 403:
		message = fmt.Sprintf("%s returned status code 403 (Forbidden)", monitor.URL)
	case statusCode == 404:
		message = fmt.Sprintf("%s returned status code 404 (Not Found)", monitor.URL)
	case statusCode == 500:
		message = fmt.Sprintf("%s returned status code 500 (Internal Server Error)", monitor.URL)
	case statusCode == 502:
		message = fmt.Sprintf("%s returned status code 502 (Bad Gateway)", monitor.URL)
	case statusCode == 503:
		message = fmt.Sprintf("%s returned status code 503 (Service Unavailable)", monitor.URL)
	case statusCode == 504:
		message = fmt.Sprintf("%s returned status code 504 (Gateway Timeout)", monitor.URL)
	default:
		message = fmt.Sprintf("%s returned status code %d", monitor.URL, statusCode)
	}

	return statusCode, message, nil
}

// executeCurlCommand executes a request using the curl command-line tool
func (s *CurlService) executeCurlCommand(monitor *types.Monitor) (int, string, error) {
	// Debug logging
	fmt.Printf("Executing curl request for URL: %s\n", monitor.URL)
	fmt.Printf("Method: %s\n", monitor.Method)
	fmt.Printf("Headers: %s\n", monitor.Headers)
	fmt.Printf("Body: %s\n", monitor.Body)

	// Build the curl command
	args := []string{
		"--location",
		"--silent",
		"--show-error",
		"-i",                      // Include headers in output
		"--connect-timeout", "10", // Add timeout
		"-v", // Verbose output for debugging
	}

	// Add method if not GET
	if monitor.Method != "GET" {
		args = append(args, "-X", monitor.Method)
	}

	// Add URL (must be after method but before other args)
	args = append(args, monitor.URL)

	// Add Content-Type header for POST/PUT requests with JSON body
	if (monitor.Method == "POST" || monitor.Method == "PUT") && monitor.Body != "" {
		// Check if the body is JSON
		if strings.HasPrefix(strings.TrimSpace(monitor.Body), "{") ||
			strings.HasPrefix(strings.TrimSpace(monitor.Body), "[") {
			args = append(args, "-H", "Content-Type: application/json")
			fmt.Printf("Added Content-Type: application/json header for JSON body\n")
		}
	}

	// Add headers
	if headers := parseHeaders(monitor.Headers); headers != nil {
		for key, value := range headers {
			args = append(args, "-H", fmt.Sprintf("%s: %s", key, value))
		}
	}

	// Add credential headers if specified
	if monitor.CredentialID != "" {
		cred, err := s.credentials.GetCredential(monitor.CredentialID)
		if err != nil {
			return 0, fmt.Sprintf("Credential retrieval failed: %v", err), err
		}

		// Get the header value based on credential type
		headerValue := s.credentials.GetHeaderValue(cred)

		// Ensure Bearer token is properly formatted
		if cred.Type == "bearer" && !strings.HasPrefix(headerValue, "Bearer ") {
			headerValue = "Bearer " + headerValue
		}

		args = append(args, "-H", fmt.Sprintf("%s: %s", cred.HeaderName, headerValue))
		fmt.Printf("Added credential header: %s: [REDACTED] (Type: %s)\n", cred.HeaderName, cred.Type)
	}

	// Add body if present
	if monitor.Body != "" {
		// Check if body is form data (starts with --form)
		if strings.Contains(monitor.Body, "--form") {
			// Parse form data from the body
			formFields := parseFormData(monitor.Body)
			for key, value := range formFields {
				args = append(args, "--form", fmt.Sprintf("%s=%s", key, value))
			}
		} else {
			// Regular data body
			args = append(args, "-d", monitor.Body)
		}
	}

	// Debug logging
	fmt.Printf("Executing curl command: curl %s\n", strings.Join(args, " "))

	// Execute the curl command
	cmd := exec.Command("curl", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = err.Error()
		}
		fmt.Printf("Curl command failed: %s\n", errMsg)
		return 0, fmt.Sprintf("Curl command failed: %s", errMsg), err
	}

	// Get the output
	output := stdout.String()
	fmt.Printf("Curl command output length: %d bytes\n", len(output))

	// Parse the HTTP status code from the response headers
	statusCode := 0
	lines := strings.Split(output, "\n")

	// Look for the status line (HTTP/1.1 200 OK)
	for _, line := range lines {
		if strings.HasPrefix(line, "HTTP/") {
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				fmt.Sscanf(parts[1], "%d", &statusCode)
				fmt.Printf("Found status code: %d from line: %s\n", statusCode, line)
				break
			}
		}
	}

	// If we couldn't parse a status code, return an error
	if statusCode == 0 {
		fmt.Printf("Failed to parse status code from curl output\n")
		return 0, "Failed to parse status code from curl output", fmt.Errorf("failed to parse status code")
	}

	// Generate appropriate message based on status code
	var message string
	switch {
	case statusCode >= 200 && statusCode < 300:
		message = fmt.Sprintf("%s returned status code %d (Success)", monitor.URL, statusCode)
	case statusCode == 401:
		message = fmt.Sprintf("%s returned status code 401 (Unauthorized)", monitor.URL)
	case statusCode == 403:
		message = fmt.Sprintf("%s returned status code 403 (Forbidden)", monitor.URL)
	case statusCode == 404:
		message = fmt.Sprintf("%s returned status code 404 (Not Found)", monitor.URL)
	case statusCode == 500:
		message = fmt.Sprintf("%s returned status code 500 (Internal Server Error)", monitor.URL)
	case statusCode == 502:
		message = fmt.Sprintf("%s returned status code 502 (Bad Gateway)", monitor.URL)
	case statusCode == 503:
		message = fmt.Sprintf("%s returned status code 503 (Service Unavailable)", monitor.URL)
	case statusCode == 504:
		message = fmt.Sprintf("%s returned status code 504 (Gateway Timeout)", monitor.URL)
	default:
		message = fmt.Sprintf("%s returned status code %d", monitor.URL, statusCode)
	}

	// Debug logging
	fmt.Printf("Status code: %d\n", statusCode)
	fmt.Printf("Message: %s\n", message)

	return statusCode, message, nil
}

// parseHeaders parses the headers string into a map
func parseHeaders(headersStr string) map[string]string {
	if headersStr == "" {
		return nil
	}

	headers := make(map[string]string)
	err := json.Unmarshal([]byte(headersStr), &headers)
	if err != nil {
		fmt.Printf("Error parsing headers: %v\n", err)
		return nil
	}

	return headers
}

// parseFormData parses form data from a string like "--form 'key=value' --form 'key2=value2'"
func parseFormData(formStr string) map[string]string {
	formData := make(map[string]string)

	fmt.Printf("Parsing form data: %s\n", formStr)

	// Split by --form
	parts := strings.Split(formStr, "--form")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		fmt.Printf("Processing form part: %s\n", part)

		// Remove quotes
		part = strings.Trim(part, "'\"")
		part = strings.TrimSpace(part)

		// Split by = to get key and value
		if idx := strings.Index(part, "="); idx > 0 {
			key := strings.TrimSpace(part[:idx])
			value := strings.TrimSpace(part[idx+1:])

			// Remove any remaining quotes
			key = strings.Trim(key, "'\"")
			value = strings.Trim(value, "'\"")

			formData[key] = value
			fmt.Printf("Added form field: %s=%s\n", key, value)
		} else {
			fmt.Printf("Could not find '=' in form part: %s\n", part)
		}
	}

	return formData
}

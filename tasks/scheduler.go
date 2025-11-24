package tasks

import (
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"sync"
	"time"
	"uptime-monitor/config"
	"uptime-monitor/repository"
	"uptime-monitor/services"
	"uptime-monitor/types"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	monitorStatusMap = make(map[string]string)    // Track previous status by monitor ID
	lastNotifiedMap  = make(map[string]time.Time) // Track last notification time for each monitor
	statusMutex      = sync.RWMutex{}
	notifyMutex      = sync.RWMutex{}
)

type Scheduler struct {
	monitors    []*types.Monitor
	mu          sync.RWMutex
	credentials *services.CredentialsService
	monitorRepo *repository.MonitorRepository
	logRepo     *repository.LogRepository
}

func NewScheduler(credentials *services.CredentialsService, monitorRepo *repository.MonitorRepository, logRepo *repository.LogRepository) *Scheduler {
	return &Scheduler{
		monitors:    make([]*types.Monitor, 0),
		credentials: credentials,
		monitorRepo: monitorRepo,
		logRepo:     logRepo,
	}
}

func (s *Scheduler) AddMonitor(monitor *types.Monitor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.monitors = append(s.monitors, monitor)
}

func (s *Scheduler) RemoveMonitor(monitorID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, m := range s.monitors {
		if m.ID == monitorID {
			s.monitors = append(s.monitors[:i], s.monitors[i+1:]...)
			break
		}
	}
}

func (s *Scheduler) Start() {
	// Log the start of the scheduler
	log.Println("🚀 SCHEDULER: Initializing Monitor Checking System 🚀")

	// Start a goroutine to periodically reload monitors from the database
	go s.periodicMonitorReload()

	// Load initial monitors
	s.reloadMonitors()

	log.Println("✅ SCHEDULER: Monitor Checking System Initialized Successfully")
}

// periodicMonitorReload reloads monitors from the database periodically
func (s *Scheduler) periodicMonitorReload() {
	// Reload monitors every minute to ensure deleted monitors are removed quickly
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("SCHEDULER: Performing periodic reload of monitors from database")
			s.reloadMonitors()
		}
	}
}

// reloadMonitors loads monitors from the repository and updates the scheduler's monitor list
func (s *Scheduler) reloadMonitors() {
	// Load monitors from repository
	log.Println("SCHEDULER: Loading monitors from repository...")
	monitors, err := s.monitorRepo.GetAllMonitors()
	if err != nil {
		log.Printf("❌ CRITICAL: Error loading monitors: %v", err)
		return
	}

	// Log total number of monitors
	log.Printf("📊 SCHEDULER: Total Monitors Loaded: %d", len(monitors))

	// Create a map of database monitors by ID for quick lookup
	dbMonitors := make(map[string]bool)
	for _, m := range monitors {
		dbMonitors[m.ID] = true
	}

	// Check if no monitors are found
	if len(monitors) == 0 {
		log.Println("⚠️ WARNING: No monitors found. Please create monitors.")

		// Clear existing monitors in the scheduler
		s.mu.Lock()
		oldCount := len(s.monitors)
		s.monitors = make([]*types.Monitor, 0)
		s.mu.Unlock()

		log.Printf("SCHEDULER: Cleared all %d monitors from scheduler", oldCount)
		return
	}

	// Create a map of existing monitors for tracking
	s.mu.RLock()
	existingMonitors := make(map[string]*types.Monitor)
	for _, m := range s.monitors {
		existingMonitors[m.ID] = m
	}
	s.mu.RUnlock()

	// Create a map of new monitors from the database
	newMonitors := make(map[string]*types.Monitor)
	for i := range monitors {
		monitor := &monitors[i]
		newMonitors[monitor.ID] = monitor
	}

	// Identify monitors to add, update, and remove
	var monitorsToAdd []*types.Monitor
	var monitorsToRemove []string

	// Find monitors to add (in new but not in existing)
	for id, monitor := range newMonitors {
		if _, exists := existingMonitors[id]; !exists {
			// Ensure monitor has a reasonable check interval
			if monitor.CheckInterval < 10 {
				log.Printf("⏱️ WARN: Monitor %s has very low check interval. Setting to 60 seconds.", monitor.Name)
				monitor.CheckInterval = 60
			}

			// Ensure monitor has an initial status
			if monitor.Status == "" {
				log.Printf("SCHEDULER: Monitor %s has no status, setting to pending", monitor.Name)
				monitor.Status = "pending"
			}

			if monitor.IsActive {
				log.Printf("🔍 SCHEDULER: Adding new monitor: %s (ID: %s, URL: %s)",
					monitor.Name, monitor.ID, monitor.URL)
				monitorsToAdd = append(monitorsToAdd, monitor)
			}
		}
	}

	// Find monitors to remove (in existing but not in new)
	for id, monitor := range existingMonitors {
		if !dbMonitors[id] {
			log.Printf("🗑️ SCHEDULER: Removing deleted monitor: %s (ID: %s)",
				monitor.Name, monitor.ID)
			monitorsToRemove = append(monitorsToRemove, id)

			// Double-check that the monitor is actually gone from the database
			// If it's not, force delete it
			var count int64
			if db, ok := s.monitorRepo.GetDB().(*gorm.DB); ok {
				db.Model(&types.Monitor{}).Where("id = ?", id).Count(&count)
				if count > 0 {
					log.Printf("WARNING: Monitor %s still exists in database after deletion! Forcing delete", id)
					db.Exec("DELETE FROM monitors WHERE id = ?", id)
				}
			}
		}
	}

	// Update the scheduler's monitor list
	s.mu.Lock()

	// Remove deleted monitors
	for _, id := range monitorsToRemove {
		for i := 0; i < len(s.monitors); i++ {
			if s.monitors[i].ID == id {
				log.Printf("Removing monitor at index %d: %s (ID: %s)",
					i, s.monitors[i].Name, s.monitors[i].ID)
				// Remove the monitor from the slice
				s.monitors = append(s.monitors[:i], s.monitors[i+1:]...)
				// Decrement i to account for the removed element
				i--
			}
		}
	}

	// Add new monitors
	for _, monitor := range monitorsToAdd {
		// Double-check that this monitor ID isn't already in our list
		isDuplicate := false
		for _, m := range s.monitors {
			if m.ID == monitor.ID {
				log.Printf("WARNING: Not adding duplicate monitor %s (ID: %s)", monitor.Name, monitor.ID)
				isDuplicate = true
				break
			}
		}

		if !isDuplicate {
			s.monitors = append(s.monitors, monitor)

			// Start monitoring for new monitors
			go func(m *types.Monitor) {
				log.Printf("🚦 SCHEDULER: Starting Monitoring for new monitor %s", m.Name)
				s.runMonitor(m)
			}(monitor)
		}
	}

	// Log the current monitors in the scheduler for debugging
	log.Printf("Current monitors in scheduler after update:")
	for i, m := range s.monitors {
		log.Printf("  %d: %s (ID: %s)", i, m.Name, m.ID)
	}

	s.mu.Unlock()

	log.Printf("SCHEDULER: Monitor list updated - Added: %d, Removed: %d, Total Active: %d",
		len(monitorsToAdd), len(monitorsToRemove), len(s.monitors))
}

func (s *Scheduler) runMonitor(monitor *types.Monitor) {
	log.Printf("MONITOR: Starting continuous monitoring for %s with interval %d seconds",
		monitor.Name, monitor.CheckInterval)

	// Perform initial check immediately
	log.Printf("MONITOR: Performing initial check for %s", monitor.Name)

	// First verify the monitor still exists in the database
	if !s.verifyMonitorExists(monitor.ID) {
		log.Printf("MONITOR %s (ID: %s) no longer exists in database, stopping monitoring",
			monitor.Name, monitor.ID)
		// Remove from scheduler's list
		s.RemoveMonitor(monitor.ID)
		return
	}

	if err := s.checkMonitor(monitor); err != nil {
		log.Printf("INITIAL CHECK ERROR for %s: %v", monitor.Name, err)
	} else {
		log.Printf("INITIAL CHECK COMPLETED for %s", monitor.Name)
	}

	// Start a ticker for periodic checks
	ticker := time.NewTicker(time.Duration(monitor.CheckInterval) * time.Second)
	defer ticker.Stop()

	// Continuous monitoring loop
	for {
		select {
		case <-ticker.C:
			log.Printf("TICKER: Time to check monitor %s", monitor.Name)

			// Verify the monitor still exists in the database before each check
			if !s.verifyMonitorExists(monitor.ID) {
				log.Printf("MONITOR %s (ID: %s) no longer exists in database, stopping monitoring",
					monitor.Name, monitor.ID)
				// Remove from scheduler's list
				s.RemoveMonitor(monitor.ID)
				return
			}

			if err := s.checkMonitor(monitor); err != nil {
				log.Printf("PERIODIC CHECK ERROR for %s: %v", monitor.Name, err)
			} else {
				log.Printf("PERIODIC CHECK COMPLETED for %s", monitor.Name)
			}
		}
	}
}

// verifyMonitorExists checks if a monitor still exists in the database
func (s *Scheduler) verifyMonitorExists(monitorID string) bool {
	_, err := s.monitorRepo.GetMonitorByID(monitorID)
	if err != nil {
		log.Printf("Monitor %s not found in database: %v", monitorID, err)
		return false
	}
	return true
}

func (s *Scheduler) checkMonitor(monitor *types.Monitor) error {
	// Add more detailed logging at the start of the method
	log.Printf("========== CHECKING MONITOR: %s ==========", monitor.Name)
	log.Printf("  URL: %s", monitor.URL)
	log.Printf("  Method: %s", monitor.Method)
	log.Printf("  Current Status: %s", monitor.Status)
	log.Printf("  Request Type: %s", monitor.RequestType)
	log.Printf("  Credential ID: %s", monitor.CredentialID)
	log.Printf("  Last Checked: %v", monitor.LastChecked)
	log.Printf("  Failure Count: %d/%d", monitor.FailureCount, monitor.FailureThreshold)

	// Explicit handling of pending state
	if monitor.Status == "" {
		log.Printf("  Setting empty status to pending")
		monitor.Status = "pending"
	}

	// More robust status determination
	previousStatus := monitor.Status
	var status string
	var message string
	var responseTime int64
	var credentialError bool = false
	var credential *services.Credential = nil

	// Check if credential exists if specified
	if monitor.CredentialID != "" {
		log.Printf("  Monitor %s uses credential ID: %s", monitor.Name, monitor.CredentialID)
		var err error
		credential, err = s.credentials.GetCredential(monitor.CredentialID)
		if err != nil {
			log.Printf("  ERROR: Failed to retrieve credential for monitor %s: %v", monitor.Name, err)
			status = "down"
			message = fmt.Sprintf("Credential error: %v", err)
			monitor.ResponseCode = 0
			monitor.FailureCount++
			credentialError = true
			log.Printf("  Credential error - Failure count increased to %d/%d",
				monitor.FailureCount, monitor.FailureThreshold)

			// Force status to down if credential error occurs and threshold is reached
			if monitor.FailureCount >= monitor.FailureThreshold {
				log.Printf("  Failure threshold reached - Setting status to DOWN")
			}
		} else {
			log.Printf("  Successfully retrieved credential: %s (%s)", credential.Name, credential.Type)
		}
	}

	// Perform monitor check only if no credential error occurred
	if !credentialError {
		// Perform monitor check
		if monitor.RequestType == "curl" {
			log.Printf("  Using CURL for request")
			curlService := services.NewCurlService(s.credentials)
			code, msg, err := curlService.ExecuteCurlRequest(monitor)
			if err != nil {
				log.Printf("  CURL ERROR: %v", err)
				status = "down"
				message = fmt.Sprintf("CURL check failed: %v", err)
				monitor.ResponseCode = 0
			} else {
				log.Printf("  CURL SUCCESS: Status code %d", code)
				monitor.ResponseCode = code
				status = getStatusFromCode(code)
				message = msg

				// Ensure we explicitly set status to "up" for successful curl requests
				if code >= 200 && code < 300 {
					log.Printf("  Setting status to UP for successful curl request")
					status = "up"
				}
			}
		} else {
			// HTTP Client check
			log.Printf("  Using HTTP client for request")
			client := &http.Client{
				Timeout: time.Duration(monitor.Timeout) * time.Second,
			}

			startTime := time.Now()
			log.Printf("  Creating request: %s %s", monitor.Method, monitor.URL)
			req, err := http.NewRequest(monitor.Method, monitor.URL, nil)
			if err != nil {
				log.Printf("  REQUEST CREATION ERROR for %s: %v", monitor.Name, err)
				status = "down"
				message = fmt.Sprintf("Request creation failed: %v", err)
				monitor.ResponseCode = 0
			} else {
				// Add headers if specified
				if headers := monitor.GetHeadersMap(); headers != nil {
					log.Printf("  Adding %d headers to request", len(headers))
					for key, value := range headers {
						req.Header.Add(key, value)
						log.Printf("    Added Header: %s: %s", key, value)
					}
				} else {
					log.Printf("  No headers to add")
				}

				// Add credential headers if specified
				if monitor.CredentialID != "" && credential != nil {
					log.Printf("  Adding credential headers")
					headerValue := s.credentials.GetHeaderValue(credential)
					req.Header.Add(credential.HeaderName, headerValue)
					log.Printf("    Added Credential Header: %s: [REDACTED]", credential.HeaderName)
				} else if monitor.CredentialID != "" && credential == nil {
					// This should not happen since we already checked for credential errors
					log.Printf("  ERROR: Credential is nil but CredentialID is set")
					status = "down"
					message = "Internal error: credential is nil"
					monitor.ResponseCode = 0
				}

				// Only proceed with request if we don't have a status yet (no credential error)
				if status == "" {
					log.Printf("  Executing HTTP request...")
					resp, err := client.Do(req)
					responseTime = time.Since(startTime).Milliseconds()
					log.Printf("  Request completed in %d ms", responseTime)

					if err != nil {
						log.Printf("  CONNECTION ERROR for %s: %v", monitor.Name, err)
						status = "down"
						message = fmt.Sprintf("Connection error: %v", err)
						monitor.ResponseCode = 0
					} else {
						defer resp.Body.Close()
						monitor.ResponseCode = resp.StatusCode
						status = getStatusFromCode(resp.StatusCode)
						message = resp.Status
						log.Printf("  Response received: %s", resp.Status)

						log.Printf("  RESPONSE DETAILS for %s:", monitor.Name)
						log.Printf("    Status Code: %d", resp.StatusCode)
						log.Printf("    Status: %s", status)
						log.Printf("    Response Time: %d ms", responseTime)
					}
				}
			}
		}

		// Enhanced status change logic
		log.Printf("  Processing status change: %s -> %s", previousStatus, status)
		if status == "down" || status == "unauthorized" {
			monitor.FailureCount++
			log.Printf("  Failure count for %s increased to %d/%d",
				monitor.Name, monitor.FailureCount, monitor.FailureThreshold)

			// Only change to down if failure threshold met
			if monitor.FailureCount >= monitor.FailureThreshold {
				log.Printf("  Failure threshold reached - Setting status to DOWN")
				status = "down"
			} else if previousStatus == "pending" {
				// Keep as pending during initial failures
				log.Printf("  Keeping status as PENDING during initial failures")
				status = "pending"
			} else {
				// Keep previous status during failure count accumulation
				log.Printf("  Keeping previous status %s during failure count accumulation", previousStatus)
				status = previousStatus
			}
		} else if status == "up" {
			// Reset failure count on successful check
			log.Printf("  Check successful - Resetting failure count")
			monitor.FailureCount = 0

			// Explicitly change from pending to up
			if previousStatus == "pending" {
				log.Printf("  Changing status from PENDING to UP")
				status = "up"
			}
		} else if status == "" && !credentialError {
			// Handle case where status wasn't set (could happen with curl)
			log.Printf("  WARNING: Status was not set, defaulting to previous status: %s", previousStatus)
			status = previousStatus
		}

		// Force transition from pending to down if we've been pending for too long
		// This ensures monitors don't get stuck in pending state
		if previousStatus == "pending" && monitor.LastChecked.Add(time.Duration(monitor.CheckInterval*3)*time.Second).Before(time.Now()) {
			log.Printf("  Monitor %s has been pending for too long, forcing to DOWN state", monitor.Name)
			status = "down"
		}

		// Update monitor status
		monitor.Status = status
		monitor.ResponseTime = responseTime
		monitor.LastChecked = time.Now()

		// Log status changes
		log.Printf("  Monitor %s status: %s -> %s (Failures: %d/%d)",
			monitor.Name,
			previousStatus,
			status,
			monitor.FailureCount,
			monitor.FailureThreshold,
		)

		// Add more detailed logging
		log.Printf("  Monitor Check Summary: %s", monitor.Name)
		log.Printf("    URL: %s", monitor.URL)
		log.Printf("    Response Code: %d", monitor.ResponseCode)
		log.Printf("    Response Time: %d ms", responseTime)
		log.Printf("    Status: %s", status)
		log.Printf("    Message: %s", message)
		log.Printf("    Last Checked: %v", monitor.LastChecked)

		// Create notification service
		log.Printf("  Checking for notification requirements")
		profileRepo := repository.NewProfileRepository(config.DB)
		methods, err := profileRepo.GetNotificationMethods(monitor.ProfileID)
		if err != nil {
			log.Printf("  Error getting notification methods: %v", err)
			return err
		}
		log.Printf("  Found %d notification methods", len(methods))

		// Convert methods to types.NotificationMethod
		var typedMethods []types.NotificationMethod
		for _, m := range methods {
			typedMethods = append(typedMethods, types.NotificationMethod{
				ID:        m.ID,
				ProfileID: m.ProfileID,
				Type:      m.Type,
				Enabled:   m.Enabled,
				Config:    m.Config,
				CreatedAt: m.CreatedAt,
				UpdatedAt: m.UpdatedAt,
			})
		}

		notificationService := services.NewNotificationService(typedMethods)

		// Check and send notification
		shouldNotify := false
		if status != previousStatus {
			// Always notify on status change
			log.Printf("  Status changed from %s to %s - Notification required", previousStatus, status)
			shouldNotify = true
		} else if status == "down" {
			// For ongoing down status, implement exponential backoff
			lastNotified, exists := lastNotifiedMap[monitor.ID]
			if !exists {
				log.Printf("  First notification for ongoing DOWN status")
				shouldNotify = true
				lastNotifiedMap[monitor.ID] = time.Now()
			} else if time.Since(lastNotified) > calculateNotificationInterval(monitor.FailureCount) {
				log.Printf("  Notification interval elapsed for ongoing DOWN status - Notification required")
				shouldNotify = true
				lastNotifiedMap[monitor.ID] = time.Now()
			} else {
				log.Printf("  Skipping notification for ongoing DOWN status - Next notification in %v",
					calculateNotificationInterval(monitor.FailureCount)-time.Since(lastNotified))
			}
		}

		if shouldNotify {
			log.Printf("  SENDING NOTIFICATION: Monitor %s changed from %s to %s",
				monitor.Name, previousStatus, status)
			err := notificationService.SendNotification(monitor, status, message)
			if err != nil {
				log.Printf("  ERROR SENDING NOTIFICATION: %v", err)
			} else {
				log.Printf("  Notification sent successfully")
			}
		} else {
			log.Printf("  No notification required")
		}

		// Update monitor in repository
		log.Printf("  Updating monitor in database")
		if err := s.monitorRepo.UpdateMonitor(monitor); err != nil {
			log.Printf("  ERROR updating monitor status: %v", err)
		} else {
			log.Printf("  Monitor updated successfully in database")
		}

		// Create log entry
		log.Printf("  Creating log entry")
		logEntry := types.Log{
			ID:        uuid.New().String(),
			MonitorID: monitor.ID,
			Status:    status,
			Message:   message,
			CreatedAt: time.Now(),
		}

		// Create log in repository
		if err := s.logRepo.CreateLog(&logEntry); err != nil {
			log.Printf("  ERROR creating log entry: %v", err)
		} else {
			log.Printf("  Log entry created successfully")
		}

		log.Printf("========== MONITOR CHECK COMPLETED: %s ==========\n", monitor.Name)
	}
	return nil
}

func getStatusFromCode(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "up"
	case code >= 300 && code < 400:
		return "redirect"
	case code >= 400 && code < 500:
		return "down"
	case code >= 500:
		return "down"
	case code == 401:
		return "unauthorized"
	default:
		return "down"
	}
}

func sendEmail(settings *types.SMTPSettings, subject, body string) error {
	log.Printf("Preparing to send email via SMTP server %s:%d to %s",
		settings.Host, settings.Port, settings.RecipientEmail)

	auth := smtp.PlainAuth("", settings.Username, settings.Password, settings.Host)
	to := []string{settings.RecipientEmail}

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n",
		settings.RecipientEmail,
		settings.Username,
		subject,
		body))

	addr := fmt.Sprintf("%s:%d", settings.Host, settings.Port)
	log.Printf("Connecting to SMTP server at %s", addr)

	err := smtp.SendMail(addr, auth, settings.Username, to, msg)
	if err != nil {
		log.Printf("SMTP Error sending to %s: %v", settings.RecipientEmail, err)
		return err
	}
	return nil
}

// Helper function for notification interval
func calculateNotificationInterval(failureCount int) time.Duration {
	// Exponential backoff for repeated notifications
	switch {
	case failureCount < 3:
		return 5 * time.Minute
	case failureCount < 10:
		return 15 * time.Minute
	case failureCount < 20:
		return 30 * time.Minute
	default:
		return 1 * time.Hour
	}
}

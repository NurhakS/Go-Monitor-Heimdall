package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"time"
	"uptime-monitor/types"
)

// NotificationService handles sending notifications through different channels
type NotificationService struct {
	methods []types.NotificationMethod
}

// NewNotificationService creates a new notification service instance
func NewNotificationService(methods []types.NotificationMethod) *NotificationService {
	return &NotificationService{methods: methods}
}

// getStatusInfo returns consistent status information across all notification types
func (s *NotificationService) getStatusInfo(status string) (symbol string, color string, displayStatus string) {
	switch strings.ToLower(status) {
	case "up":
		return "✅", "good", "UP"
	case "down":
		return "❌", "danger", "DOWN"
	case "unauthorized", "401":
		return "⚠️", "warning", "UNAUTHORIZED"
	default:
		return "⏳", "default", strings.ToUpper(status)
	}
}

// SendNotification sends notifications through all enabled channels
func (s *NotificationService) SendNotification(monitor *types.Monitor, status, message string) error {
	var errs []error
	notificationSent := make(map[string]bool)

	for _, method := range s.methods {
		// Skip if this notification type was already sent
		if notificationSent[method.Type] {
			continue
		}

		// Parse config JSON
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(method.Config), &config); err != nil {
			errs = append(errs, fmt.Errorf("config parse error for %s: %v", method.Type, err))
			continue
		}

		switch method.Type {
		case "email":
			if err := s.SendEmail(monitor, status, message, config); err != nil {
				errs = append(errs, fmt.Errorf("email error: %v", err))
			}
			notificationSent[method.Type] = true
		case "slack":
			if err := s.SendSlack(monitor, status, message, config); err != nil {
				errs = append(errs, fmt.Errorf("slack error: %v", err))
			}
			notificationSent[method.Type] = true
		case "teams":
			if err := s.SendTeams(monitor, status, message, config); err != nil {
				errs = append(errs, fmt.Errorf("teams error: %v", err))
			}
			notificationSent[method.Type] = true
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("notification errors: %v", errs)
	}
	return nil
}

// SendEmail sends notification via email
func (s *NotificationService) SendEmail(monitor *types.Monitor, status, message string, config map[string]interface{}) error {
	symbol, _, displayStatus := s.getStatusInfo(status)

	subject := fmt.Sprintf("%s Monitor Alert: %s is %s", symbol, monitor.Name, displayStatus)
	body := fmt.Sprintf(
		"Monitor Status Notification\n\n"+
			"%s Status: %s\n"+
			"Monitor: %s\n"+
			"URL: %s\n"+
			"Message: %s\n"+
			"Response Code: %d\n"+
			"Response Time: %dms\n"+
			"Time: %s",
		symbol,
		displayStatus,
		monitor.Name,
		monitor.URL,
		message,
		monitor.ResponseCode,
		monitor.ResponseTime,
		time.Now().Format(time.RFC1123),
	)

	smtpHost := config["smtp_host"].(string)
	smtpPort := config["smtp_port"].(float64)
	smtpEmail := config["smtp_email"].(string)
	smtpPassword := config["smtp_password"].(string)
	recipientEmail := config["recipient_email"].(string)

	auth := smtp.PlainAuth("", smtpEmail, smtpPassword, smtpHost)
	to := []string{recipientEmail}

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n",
		recipientEmail,
		smtpEmail,
		subject,
		body))

	addr := fmt.Sprintf("%s:%v", smtpHost, int(smtpPort))
	return smtp.SendMail(addr, auth, smtpEmail, to, msg)
}

// SendSlack sends notification to Slack
func (s *NotificationService) SendSlack(monitor *types.Monitor, status, message string, config map[string]interface{}) error {
	symbol, _, displayStatus := s.getStatusInfo(status)
	webhookURL := config["webhook_url"].(string)
	channel := config["channel"].(string)

	payload := map[string]interface{}{
		"channel": channel,
		"blocks": []map[string]interface{}{
			{
				"type": "header",
				"text": map[string]interface{}{
					"type":  "plain_text",
					"text":  fmt.Sprintf("%s Monitor Alert: %s", symbol, monitor.Name),
					"emoji": true,
				},
			},
			{
				"type": "section",
				"fields": []map[string]interface{}{
					{
						"type": "mrkdwn",
						"text": fmt.Sprintf("*Status:*\n%s %s", symbol, displayStatus),
					},
					{
						"type": "mrkdwn",
						"text": fmt.Sprintf("*URL:*\n%s", monitor.URL),
					},
					{
						"type": "mrkdwn",
						"text": fmt.Sprintf("*Message:*\n%s", message),
					},
					{
						"type": "mrkdwn",
						"text": fmt.Sprintf("*Response Code:*\n%d", monitor.ResponseCode),
					},
					{
						"type": "mrkdwn",
						"text": fmt.Sprintf("*Response Time:*\n%dms", monitor.ResponseTime),
					},
				},
			},
			{
				"type": "context",
				"elements": []map[string]interface{}{
					{
						"type": "mrkdwn",
						"text": fmt.Sprintf("Time: %s", time.Now().Format(time.RFC1123)),
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack API returned status code: %d", resp.StatusCode)
	}

	return nil
}

// SendTeams sends notification to Microsoft Teams
func (s *NotificationService) SendTeams(monitor *types.Monitor, status, message string, config map[string]interface{}) error {
	symbol, cardColor, displayStatus := s.getStatusInfo(status)
	webhookURL := config["webhook_url"].(string)

	title := fmt.Sprintf("%s Monitor Alert: %s", symbol, monitor.Name)

	payload := map[string]interface{}{
		"type": "message",
		"attachments": []map[string]interface{}{
			{
				"contentType": "application/vnd.microsoft.card.adaptive",
				"content": map[string]interface{}{
					"type": "AdaptiveCard",
					"body": []map[string]interface{}{
						{
							"type":   "TextBlock",
							"text":   title,
							"weight": "bolder",
							"size":   "large",
							"wrap":   true,
						},
						{
							"type":   "TextBlock",
							"text":   fmt.Sprintf("Status: %s %s", symbol, displayStatus),
							"weight": "bolder",
							"color":  cardColor,
							"wrap":   true,
						},
						{
							"type": "TextBlock",
							"text": fmt.Sprintf("URL: %s\nMessage: %s\nResponse Code: %d\nResponse Time: %dms\nTime: %s",
								monitor.URL,
								message,
								monitor.ResponseCode,
								monitor.ResponseTime,
								time.Now().Format(time.RFC1123)),
							"wrap": true,
						},
					},
					"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
					"version": "1.2",
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Teams API returned status code: %d", resp.StatusCode)
	}

	return nil
}

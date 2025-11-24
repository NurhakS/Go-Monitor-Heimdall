package utils

import (
	"fmt"
	"net/smtp"
)

func SendEmail(smtpEmail, smtpPassword, smtpServer, smtpPort, to, subject, body string) error {
	// Create the email message
	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s", smtpEmail, to, subject, body)

	// Setup SMTP authentication
	auth := smtp.PlainAuth("", smtpEmail, smtpPassword, smtpServer)

	// Send the email
	return smtp.SendMail(smtpServer+":"+smtpPort, auth, smtpEmail, []string{to}, []byte(msg))
}

package mail

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"os"
)

type Mail struct {
	Username string
	Password string
	Host     string
	Port     string
}

func NewMail() *Mail {
	return &Mail{
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
	}
}

// SendMail sends an email to the provided recipient with the given subject and message.
// The sender is identified by the SMTP_USERNAME, SMTP_PASSWORD, SMTP_HOST, and SMTP_PORT environment variables.
func (m *Mail) SendMail(to, subject, message string) error {
	auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)

	// Include the subject and From headers
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		m.Username, to, subject, message)

	err := smtp.SendMail(m.Host+":"+m.Port, auth, m.Username, []string{to}, []byte(msg))
	if err != nil {
		slog.Error("Error sending email", "error", err)
		return err
	}

	slog.Info("Email sent successfully", "recipient", to)

	return nil
}

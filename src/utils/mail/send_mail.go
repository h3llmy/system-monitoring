package mail

import (
	"log"
	"net/smtp"
	"os"
)

// SendMail sends a mail to the provided recipient with the given subject and message.
// The sender is identified by the SMTP_USERNAME, SMTP_PASSWORD, SMTP_HOST, and SMTP_PORT environment variables.
// This function will log and exit if an error occurs while sending the mail.
func SendMail(to, subject, message string) {
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	log.Printf("username: %s, password: %s, host: %s, port: %s", username, password, host, port)
	log.Printf("Sending email to %s with subject %s and message %s", to, subject, message)
	auth := smtp.PlainAuth("", username, password, host)
	err := smtp.SendMail(host+":"+port, auth, username, []string{to}, []byte(message))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Email sent successfully")
}

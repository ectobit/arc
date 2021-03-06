// Package smtp contains implementation of message sender using SMTP server.
package smtp

import (
	"fmt"
	"net/smtp"

	"go.ectobit.com/arc/send"
	"go.ectobit.com/lax"
)

var _ send.Sender = (*Mailer)(nil)

// Mailer implements send.Sender interface using SMTP server.
type Mailer struct {
	smtpHost string
	smtpPort uint16
	username string
	password string
	sender   string
	log      lax.Logger
}

// NewMailer creates mailer.
func NewMailer(smtpHost string, smtpPort uint16, username, password, sender string, log lax.Logger) *Mailer {
	return &Mailer{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		username: username,
		password: password,
		sender:   sender,
		log:      log,
	}
}

// Send sends message using SMTP server.
func (m *Mailer) Send(recipient, subject, message string) error {
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", m.sender, recipient, subject, message))
	auth := smtp.PlainAuth("", m.username, m.password, m.smtpHost)
	server := fmt.Sprintf("%s:%d", m.smtpHost, m.smtpPort)

	m.log.Info("send mail", lax.String("server", server), lax.String("recipient", recipient))

	if err := smtp.SendMail(server, auth, m.sender, []string{recipient}, msg); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}

	return nil
}

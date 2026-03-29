package notifier

import (
	"context"
	"fmt"
	"sync"

	mail "github.com/wneessen/go-mail"
)

// EmailSender implements the Sender interface for sending emails via SMTP.
// It maintains a persistent SMTP connection that is re-established if it fails.
type EmailSender struct {
	mu     sync.Mutex
	config EmailConfig
	client *mail.Client
}

// EmailConfig contains the configuration for the EmailSender.
type EmailConfig struct {
	// Host is the SMTP server hostname.
	Host string
	// Port is the SMTP server port (e.g., 587).
	Port int
	// User is the SMTP username (usually the email address).
	User string
	// Password is the SMTP password.
	Password string
	// FromName is the name that will appear as the sender.
	FromName string
	// DisableTLS disables encryption for the SMTP connection.
	// When true, it uses unencrypted authentication and no TLS/STARTTLS.
	// This is typically used for local development or internal relays that do not support or require encryption.
	DisableTLS bool
}

// NewEmailSender creates a new EmailSender with the provided configuration.
func NewEmailSender(config EmailConfig) *EmailSender {
	return &EmailSender{
		config: config,
	}
}

// Type returns the name of the sender type ("Email").
func (s *EmailSender) Type() string {
	return "Email"
}

// Send sends an email notification to the specified contact.
func (s *EmailSender) Send(_ context.Context, level Level, contact *Contact, subject, message string) error {
	if contact.Email == "" {
		return nil
	}

	client, err := s.getClient()
	if err != nil {
		return fmt.Errorf("failed to get SMTP client: %w", err)
	}

	m := mail.NewMsg()
	if s.config.FromName != "" {
		if err := m.FromFormat(s.config.FromName, s.config.User); err != nil {
			return fmt.Errorf("failed to set from address: %w", err)
		}
	} else {
		if err := m.From(s.config.User); err != nil {
			return fmt.Errorf("failed to set from address: %w", err)
		}
	}

	if err := m.To(contact.Email); err != nil {
		return fmt.Errorf("failed to add recipient: %w", err)
	}

	if subject == "" {
		subject = fmt.Sprintf("[%s] Notification", level.String())
	}
	m.Subject(subject)
	m.SetBodyString(mail.TypeTextPlain, message)

	if err := client.DialAndSend(m); err != nil {
		// If sending fails, it might be a broken connection, reset the client
		s.resetClient()
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *EmailSender) getClient() (*mail.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if client exists and is alive
	if s.client != nil {
		return s.client, nil
	}

	opts := []mail.Option{
		mail.WithPort(s.config.Port),
		mail.WithUsername(s.config.User),
		mail.WithPassword(s.config.Password),
	}
	if s.config.DisableTLS {
		opts = append(opts, mail.WithSMTPAuth(mail.SMTPAuthPlainNoEnc), mail.WithTLSPolicy(mail.NoTLS))
	} else {
		opts = append(opts, mail.WithSMTPAuth(mail.SMTPAuthPlain), mail.WithTLSPolicy(mail.TLSOpportunistic))
	}

	client, err := mail.NewClient(s.config.Host, opts...)
	if err != nil {
		return nil, err
	}

	s.client = client
	return s.client, nil
}

func (s *EmailSender) resetClient() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client != nil {
		s.client = nil
	}
}

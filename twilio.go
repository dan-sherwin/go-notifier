package notifier

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

// TwilioSender implements the Sender interface for sending SMS messages via Twilio.
type TwilioSender struct {
	config TwilioConfig
	client *twilio.RestClient
}

// TwilioConfig contains the configuration for the TwilioSender.
type TwilioConfig struct {
	// AccountSid is your Twilio account SID.
	AccountSid string
	// AuthToken is your Twilio authentication token.
	AuthToken string
	// From is the Twilio phone number (in E.164 format) or an Alphanumeric Sender ID.
	From string
}

// NewTwilioSender creates a new TwilioSender with the provided configuration.
func NewTwilioSender(config TwilioConfig) *TwilioSender {
	return &TwilioSender{
		config: config,
		client: twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: config.AccountSid,
			Password: config.AuthToken,
		}),
	}
}

// Type returns the name of the sender type ("SMS").
func (s *TwilioSender) Type() string {
	return "SMS"
}

// Send sends an SMS notification to the specified contact via Twilio.
func (s *TwilioSender) Send(_ context.Context, _ Level, contact *Contact, subject, message string) error {
	if contact.SMS == "" {
		return nil
	}

	body := subject
	if message != "" {
		if len(body) > 0 {
			body += "\n\n"
		}
		body += message
	}

	if body == "" {
		return fmt.Errorf("empty notification body")
	}

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(s.cleanNumber(contact.SMS))
	params.SetFrom(s.config.From)
	params.SetBody(body)

	_, err := s.client.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	return nil
}

func (s *TwilioSender) cleanNumber(number string) string {
	re := regexp.MustCompile(`\D`)
	clean := strings.TrimSpace(re.ReplaceAllString(number, ""))
	if !strings.HasPrefix(clean, "+") {
		// Defaulting to US prefix if not present, but better would be full libphonenumber support.
		// For now keeping similar logic but ensuring "+"
		if !strings.HasPrefix(clean, "1") && len(clean) == 10 {
			clean = "+1" + clean
		} else if !strings.HasPrefix(clean, "+") {
			clean = "+" + clean
		}
	}
	return clean
}

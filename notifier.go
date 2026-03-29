package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// Level defines the severity of the notification.
type Level int

const (
	// LevelInfo defines an informational notification level.
	LevelInfo Level = iota
	// LevelWarn defines a warning notification level.
	LevelWarn
	// LevelCrit defines a critical notification level.
	LevelCrit
)

// String returns the string representation of the Level.
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelCrit:
		return "CRIT"
	default:
		return "UNKNOWN"
	}
}

// Sender is the interface for notification providers like Email, SMS, or Slack.
// Implementations must be thread-safe as they may be called from multiple goroutines.
type Sender interface {
	// Send sends a notification to the specified contact.
	Send(ctx context.Context, level Level, contact *Contact, subject, message string) error
	// Type returns the name of the sender type (e.g., "Email", "SMS").
	Type() string
}

// Notifier handles sending notifications across multiple channels to contacts.
// It supports multiple senders and manages contact-specific preferences.
type Notifier struct {
	mu       sync.RWMutex
	contacts map[string]*Contact
	senders  []Sender
	logger   *slog.Logger
	wg       sync.WaitGroup
}

// Option is a functional option for configuring the Notifier.
type Option func(*Notifier)

// New creates a new Notifier with the provided options.
func New(opts ...Option) *Notifier {
	n := &Notifier{
		contacts: make(map[string]*Contact),
		logger:   slog.Default(),
	}

	for _, opt := range opts {
		opt(n)
	}

	return n
}

// WithLogger sets the logger for the Notifier.
func WithLogger(logger *slog.Logger) Option {
	return func(n *Notifier) {
		if logger != nil {
			n.logger = logger
		}
	}
}

// WithSender adds a notification sender (e.g., Email, SMS) to the Notifier.
// Multiple senders can be added to send notifications via different channels.
func WithSender(sender Sender) Option {
	return func(n *Notifier) {
		if sender != nil {
			n.senders = append(n.senders, sender)
		}
	}
}

// SetContact adds or updates a contact in the Notifier.
func (n *Notifier) SetContact(id string, contact *Contact) error {
	if id == "" || contact == nil {
		return fmt.Errorf("invalid contact: ID and contact must not be empty")
	}

	n.mu.Lock()
	defer n.mu.Unlock()
	n.contacts[id] = contact
	return nil
}

// Info sends an informational notification.
func (n *Notifier) Info(ctx context.Context, subject, message string) {
	n.Notify(ctx, LevelInfo, subject, message)
}

// Warn sends a warning notification.
func (n *Notifier) Warn(ctx context.Context, subject, message string) {
	n.Notify(ctx, LevelWarn, subject, message)
}

// Crit sends a critical notification.
func (n *Notifier) Crit(ctx context.Context, subject, message string) {
	n.Notify(ctx, LevelCrit, subject, message)
}

// Notify sends a notification to all contacts matching the level.
func (n *Notifier) Notify(ctx context.Context, level Level, subject, message string) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, c := range n.contacts {
		if !c.ShouldNotify(level) {
			continue
		}

		for _, sender := range n.senders {
			n.wg.Add(1)
			go func(s Sender, contact *Contact) {
				defer n.wg.Done()
				if err := s.Send(ctx, level, contact, subject, message); err != nil {
					n.logger.Error("failed to send notification",
						"type", s.Type(),
						"contact", contact.Name,
						"level", level.String(),
						"error", err)
				}
			}(sender, c)
		}
		c.MarkNotified()
	}
}

// Wait blocks until all pending notifications are sent.
func (n *Notifier) Wait() {
	n.wg.Wait()
}

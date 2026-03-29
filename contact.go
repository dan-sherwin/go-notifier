// Package notifier provides a flexible and thread-safe notification system
// with support for multiple senders (Email, SMS) and contact-specific throttling.
package notifier

import (
	"sync"
	"time"
)

// Levels defines which notification severities are enabled for a contact.
type Levels struct {
	// Info indicates if informational notifications are enabled.
	Info bool
	// Warn indicates if warning notifications are enabled.
	Warn bool
	// Crit indicates if critical notifications are enabled.
	Crit bool
}

// Contact contains recipient information and notification preferences.
// It tracks the last notification time to implement throttling.
type Contact struct {
	mu sync.RWMutex

	// Name is the display name of the contact.
	Name string
	// Email is the recipient's email address.
	Email string
	// SMS is the recipient's phone number for SMS.
	SMS string
	// Levels defines which notification levels are enabled for this contact.
	Levels Levels
	// MinimumNotificationInterval defines the minimum time that must pass between notifications.
	MinimumNotificationInterval time.Duration
	// LastNotification is the timestamp when the last notification was sent to this contact.
	LastNotification time.Time
}

// ShouldNotify checks if a notification should be sent based on level and interval.
func (c *Contact) ShouldNotify(level Level) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if the level is enabled for this contact
	enabled := false
	switch level {
	case LevelInfo:
		enabled = c.Levels.Info
	case LevelWarn:
		enabled = c.Levels.Warn
	case LevelCrit:
		enabled = c.Levels.Crit
	}

	if !enabled {
		return false
	}

	// Check if enough time has passed since the last notification
	if c.LastNotification.IsZero() || c.MinimumNotificationInterval == 0 {
		return true
	}

	return time.Since(c.LastNotification) >= c.MinimumNotificationInterval
}

// MarkNotified updates the last notification timestamp for the contact.
func (c *Contact) MarkNotified() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastNotification = time.Now()
}

package notifier

import (
	"context"
	"log/slog"
)

var (
	defaultNotifier *Notifier
)

// Default returns the default global notifier instance.
func Default() *Notifier {
	if defaultNotifier == nil {
		// Note: By default it has no senders and no contacts.
		// It's up to the user to configure it via InitDefault.
		defaultNotifier = New()
	}
	return defaultNotifier
}

// InitDefault initializes the default notifier with the provided options.
func InitDefault(opts ...Option) {
	defaultNotifier = New(opts...)
}

// Info sends an informational notification via the default notifier.
func Info(ctx context.Context, subject, message string) {
	Default().Info(ctx, subject, message)
}

// Warn sends a warning notification via the default notifier.
func Warn(ctx context.Context, subject, message string) {
	Default().Warn(ctx, subject, message)
}

// Crit sends a critical notification via the default notifier.
func Crit(ctx context.Context, subject, message string) {
	Default().Crit(ctx, subject, message)
}

// SetContact adds or updates a contact in the default notifier.
func SetContact(id string, contact *Contact) error {
	return Default().SetContact(id, contact)
}

// SetLogger sets the logger for the default notifier.
func SetLogger(logger *slog.Logger) {
	Default().mu.Lock()
	defer Default().mu.Unlock()
	Default().logger = logger
}

// Wait blocks until all pending notifications are sent by the default notifier.
func Wait() {
	Default().Wait()
}

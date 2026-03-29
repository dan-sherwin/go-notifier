# go-notifier

[![Go Reference](https://pkg.go.dev/badge/github.com/dan-sherwin/go-notifier.svg)](https://pkg.go.dev/github.com/dan-sherwin/go-notifier)
[![Go Report Card](https://goreportcard.com/badge/github.com/dan-sherwin/go-notifier)](https://goreportcard.com/report/github.com/dan-sherwin/go-notifier)

`go-notifier` is a modern, thread-safe, and asynchronous notification SDK for Go. It allows you to send notifications via multiple channels (Email, SMS) to multiple contacts with built-in throttling and severity levels.

## Features

- **Asynchronous Execution:** Notifications are sent in background goroutines to prevent blocking your application.
- **Multiple Channels:** Built-in support for SMTP Email and Twilio SMS.
- **Thread-Safe:** Designed for concurrent use in high-traffic applications.
- **Throttling:** Per-contact `MinimumNotificationInterval` to prevent notification fatigue.
- **Severity Levels:** Support for `LevelInfo`, `LevelWarn`, and `LevelCrit` to filter notifications.
- **Flexible Configuration:** Uses functional options for easy setup and dependency injection.
- **TLS Control:** Ability to disable TLS for unencrypted connections to local relays or for development.
- **Context Support:** Fully supports `context.Context` for cancellation and timeouts.

## Installation

```bash
go get github.com/dan-sherwin/go-notifier
```

## Quick Start

The simplest way to use `go-notifier` is via the global default notifier.

```go
package main

import (
	"context"
	"time"

	"github.com/dan-sherwin/go-notifier"
)

func main() {
	ctx := context.Background()

	// 1. Initialize the default notifier with senders
	notifier.InitDefault(
		notifier.WithSender(notifier.NewEmailSender(notifier.EmailConfig{
			Host:     "smtp.example.com",
			Port:     587,
			User:     "alerts@example.com",
			Password: "your-password",
			FromName: "System Alerts",
			// DisableTLS: true, // Use for local relays that do not require encryption
		})),
	)

	// 2. Add contacts
	notifier.SetContact("admin", &notifier.Contact{
		Name:  "Admin User",
		Email: "admin@example.com",
		Levels: notifier.Levels{
			Info: true,
			Warn: true,
			Crit: true,
		},
		MinimumNotificationInterval: 5 * time.Minute,
	})

	// 3. Send notifications
	notifier.Info(ctx, "System Started", "The application has successfully started.")
	
	// Ensure all background notifications are sent before exiting
	notifier.Wait()
}
```

## Advanced Usage

### Custom Notifier Instance

You can create multiple independent `Notifier` instances instead of using the global one.

```go
n := notifier.New(
    notifier.WithLogger(slog.Default()),
    notifier.WithSender(emailSender),
    notifier.WithSender(twilioSender),
)

n.SetContact("user-1", contact)
n.Notify(ctx, notifier.LevelCrit, "Database Down", "Unable to connect to production DB.")
n.Wait()
```

### SMS via Twilio

```go
twilioSender := notifier.NewTwilioSender(notifier.TwilioConfig{
    AccountSid: "ACxxxxxxxxxxxxxxxxxxxxxxxx",
    AuthToken:  "your_auth_token",
    From:       "+1234567890",
})

n := notifier.New(notifier.WithSender(twilioSender))
```

### Throttling (Rate Limiting)

Each contact can have a `MinimumNotificationInterval`. If multiple notifications are sent within this window, only the first one will be delivered.

```go
contact := &notifier.Contact{
    Name: "Dan",
    Email: "dan@example.com",
    MinimumNotificationInterval: 1 * time.Hour, // Only receive 1 email per hour
    Levels: notifier.Levels{Crit: true},
}
```

## Development

### Running Tests

```bash
go test ./...
```

### Linting

This project uses `golangci-lint`. You can run it locally using:

```bash
golangci-lint run
```

## License

[MIT](LICENSE)

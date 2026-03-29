package notifier

import (
	"context"
	"testing"
	"time"
)

type mockSender struct {
	sentCount int
}

func (m *mockSender) Send(_ context.Context, _ Level, _ *Contact, _, _ string) error {
	m.sentCount++
	return nil
}

func (m *mockSender) Type() string {
	return "Mock"
}

func TestNotifier(t *testing.T) {
	mock := &mockSender{}
	n := New(WithSender(mock))

	contact := &Contact{
		Name:  "Test User",
		Email: "test@example.com",
		Levels: Levels{
			Info: true,
			Warn: true,
		},
	}

	_ = n.SetContact("test", contact)

	ctx := context.Background()
	n.Info(ctx, "Test Subject", "Test Message")
	n.Wait()

	if mock.sentCount != 1 {
		t.Errorf("expected 1 notification sent, got %d", mock.sentCount)
	}

	// Should not notify because of level
	n.Crit(ctx, "Crit Subject", "Crit Message")
	n.Wait()
	if mock.sentCount != 1 {
		t.Errorf("expected 1 notification sent after Crit, got %d", mock.sentCount)
	}

	// Test interval
	contact.MinimumNotificationInterval = 0
	n.Warn(ctx, "Warn Subject", "Warn Message")
	n.Wait()
	if mock.sentCount != 2 {
		t.Errorf("expected 2 notifications sent after Warn, got %d", mock.sentCount)
	}

	contact.MinimumNotificationInterval = 1 * time.Hour
	n.Warn(ctx, "Warn Subject 2", "Warn Message 2")
	n.Wait()
	if mock.sentCount != 2 {
		t.Errorf("expected 2 notifications sent after throttled Warn, got %d", mock.sentCount)
	}
}

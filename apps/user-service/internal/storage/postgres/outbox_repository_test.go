package postgres

import (
	"context"
	"testing"

	"gorm.io/gorm"

	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/domain"
)

// seedUser creates a real user (and therefore a real "user.created"
// outbox event, via Create's transaction) for outbox-repository tests to
// work against, rather than inserting outbox rows directly — that keeps
// these tests exercising the same path production actually takes.
func seedUser(t *testing.T, db *gorm.DB, id, kimoID, phoneNumber string) {
	t.Helper()
	repo := NewUserRepository(db)
	if _, err := repo.Create(context.Background(), id, kimoID, domain.RegisterInput{
		PhoneNumber: phoneNumber,
		FullName:    "Outbox Test User",
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func TestOutboxRepository_FetchUnpublished(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, "11111111-1111-4111-8111-111111111111", "ABCDEF123456", "+6281234567890")
	seedUser(t, db, "22222222-2222-4222-8222-222222222222", "GHIJKL789012", "+6289876543210")

	repo := NewOutboxRepository(db)

	events, err := repo.FetchUnpublished(context.Background(), 100)
	if err != nil {
		t.Fatalf("FetchUnpublished() error = %v, want nil", err)
	}
	if len(events) != 2 {
		t.Fatalf("got %d unpublished events, want 2", len(events))
	}
	for _, event := range events {
		if event.EventType != EventTypeUserCreated {
			t.Fatalf("got event_type %q, want %q", event.EventType, EventTypeUserCreated)
		}
	}
}

func TestOutboxRepository_FetchUnpublished_RespectsLimit(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, "11111111-1111-4111-8111-111111111111", "ABCDEF123456", "+6281234567890")
	seedUser(t, db, "22222222-2222-4222-8222-222222222222", "GHIJKL789012", "+6289876543210")

	repo := NewOutboxRepository(db)

	events, err := repo.FetchUnpublished(context.Background(), 1)
	if err != nil {
		t.Fatalf("FetchUnpublished() error = %v, want nil", err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d events, want exactly 1 (the limit)", len(events))
	}
}

func TestOutboxRepository_MarkPublished(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, "11111111-1111-4111-8111-111111111111", "ABCDEF123456", "+6281234567890")

	repo := NewOutboxRepository(db)
	ctx := context.Background()

	before, err := repo.FetchUnpublished(ctx, 100)
	if err != nil {
		t.Fatalf("FetchUnpublished() error = %v, want nil", err)
	}
	if len(before) != 1 {
		t.Fatalf("got %d unpublished events before MarkPublished, want 1", len(before))
	}

	if err := repo.MarkPublished(ctx, before[0].ID); err != nil {
		t.Fatalf("MarkPublished() error = %v, want nil", err)
	}

	after, err := repo.FetchUnpublished(ctx, 100)
	if err != nil {
		t.Fatalf("FetchUnpublished() error = %v, want nil", err)
	}
	if len(after) != 0 {
		t.Fatalf("got %d unpublished events after MarkPublished, want 0", len(after))
	}
}

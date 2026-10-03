package postgres

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/domain"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/idgen"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/migrations"
)

// newTestDB connects to a real Postgres instance for integration testing.
// Skips (does not fail) when TRANSACTION_SERVICE_TEST_DATABASE_URL is
// unset — same convention as user-service's/wallet-service's newTestDB.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	databaseURL := os.Getenv("TRANSACTION_SERVICE_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TRANSACTION_SERVICE_TEST_DATABASE_URL not set; skipping Postgres integration test")
	}

	db, err := Open(databaseURL)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get underlying sql.DB: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := Migrate(ctx, sqlDB, migrations.Files); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	if err := db.Exec("TRUNCATE TABLE transactions, outbox_events").Error; err != nil {
		t.Fatalf("truncate transactions/outbox_events tables before test: %v", err)
	}

	return db
}

func testRequest() domain.TransactionRequest {
	return domain.TransactionRequest{
		IdempotencyKey:   "intent-123",
		SenderUserID:     "11111111-1111-4111-8111-111111111111",
		SenderWalletID:   "22222222-2222-4222-8222-222222222222",
		ReceiverUserID:   "33333333-3333-4333-8333-333333333333",
		ReceiverWalletID: "44444444-4444-4444-8444-444444444444",
		Currency:         "IDR",
		Amount:           5000,
	}
}

func mustNewID(t *testing.T) string {
	t.Helper()
	id, err := idgen.NewV4()
	if err != nil {
		t.Fatalf("generate id: %v", err)
	}
	return id
}

func TestTransactionRepository_Create(t *testing.T) {
	db := newTestDB(t)
	repo := NewTransactionRepository(db)
	ctx := context.Background()

	request := testRequest()
	id := mustNewID(t)

	created, err := repo.Create(ctx, id, request)
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if created.ID != id {
		t.Fatalf("Create() returned id %q, want %q", created.ID, id)
	}
	if created.Status != statusPending {
		t.Fatalf("Create() returned status %q, want %q", created.Status, statusPending)
	}
	if created.SenderWalletID != request.SenderWalletID || created.ReceiverWalletID != request.ReceiverWalletID {
		t.Fatalf("Create() returned unexpected transaction: %+v", created)
	}
	if created.CreatedAt.IsZero() {
		t.Fatalf("Create() returned zero CreatedAt")
	}

	// The transactional outbox write is the whole point of Create's
	// transaction — prove the event actually landed, not just the row.
	var outboxRow outboxEventModel
	if err := db.Where("event_type = ?", EventTypeTransactionCreated).First(&outboxRow).Error; err != nil {
		t.Fatalf("expected a %q outbox event, got error: %v", EventTypeTransactionCreated, err)
	}
	if outboxRow.PublishedAt != nil {
		t.Fatalf("outbox event was already marked published at insert time, want nil until the relay publishes it")
	}
	if len(outboxRow.Payload) == 0 {
		t.Fatalf("outbox event has an empty payload")
	}
}

func TestTransactionRepository_Create_DuplicateIdempotencyKey(t *testing.T) {
	db := newTestDB(t)
	repo := NewTransactionRepository(db)
	ctx := context.Background()

	request := testRequest()

	if _, err := repo.Create(ctx, mustNewID(t), request); err != nil {
		t.Fatalf("first Create() error = %v, want nil", err)
	}

	_, err := repo.Create(ctx, mustNewID(t), request)
	if !errors.Is(err, domain.ErrIdempotencyConflict) {
		t.Fatalf("second Create() error = %v, want %v", err, domain.ErrIdempotencyConflict)
	}

	// The failed second Create() must not have left behind an outbox
	// event — the whole transaction, row insert AND outbox insert
	// together, must have rolled back.
	var outboxCount int64
	if err := db.Model(&outboxEventModel{}).Count(&outboxCount).Error; err != nil {
		t.Fatalf("count outbox events: %v", err)
	}
	if outboxCount != 1 {
		t.Fatalf("got %d outbox events after one success and one rejected duplicate, want exactly 1", outboxCount)
	}
}

// TestTransactionRepository_Create_ConcurrentDuplicateIdempotencyKey proves
// the uniqueness guarantee holds under a race, not just sequential calls —
// two goroutines submit the same (idempotency_key, sender_user_id) at
// once; exactly one must succeed.
func TestTransactionRepository_Create_ConcurrentDuplicateIdempotencyKey(t *testing.T) {
	db := newTestDB(t)
	repo := NewTransactionRepository(db)
	ctx := context.Background()

	request := testRequest()
	ids := []string{mustNewID(t), mustNewID(t)}

	var wg sync.WaitGroup
	errs := make([]error, len(ids))
	for i, id := range ids {
		wg.Add(1)
		go func(i int, id string) {
			defer wg.Done()
			_, errs[i] = repo.Create(ctx, id, request)
		}(i, id)
	}
	wg.Wait()

	successes, conflicts := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrIdempotencyConflict):
			conflicts++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("got %d successes and %d conflicts, want exactly 1 and 1", successes, conflicts)
	}

	var rowCount int64
	if err := db.Model(&transactionModel{}).Where("idempotency_key = ? AND sender_user_id = ?", request.IdempotencyKey, request.SenderUserID).Count(&rowCount).Error; err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("got %d transaction rows for one idempotency key, want 1", rowCount)
	}
}

func TestTransactionRepository_UpdateStatus_Completed(t *testing.T) {
	db := newTestDB(t)
	repo := NewTransactionRepository(db)
	ctx := context.Background()

	id := mustNewID(t)
	if _, err := repo.Create(ctx, id, testRequest()); err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	if err := repo.UpdateStatus(ctx, id, "COMPLETED", nil); err != nil {
		t.Fatalf("UpdateStatus() error = %v, want nil", err)
	}

	var row transactionModel
	if err := db.Where("id = ?", id).First(&row).Error; err != nil {
		t.Fatalf("fetch updated row: %v", err)
	}
	if row.Status != "COMPLETED" {
		t.Fatalf("status = %q, want %q", row.Status, "COMPLETED")
	}
	if row.FailureReason != nil {
		t.Fatalf("failure_reason = %v, want nil", *row.FailureReason)
	}
	if row.CompletedAt == nil {
		t.Fatalf("completed_at is nil, want set")
	}
}

func TestTransactionRepository_UpdateStatus_FailedWithReason(t *testing.T) {
	db := newTestDB(t)
	repo := NewTransactionRepository(db)
	ctx := context.Background()

	id := mustNewID(t)
	if _, err := repo.Create(ctx, id, testRequest()); err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	reason := "insufficient balance"
	if err := repo.UpdateStatus(ctx, id, "FAILED", &reason); err != nil {
		t.Fatalf("UpdateStatus() error = %v, want nil", err)
	}

	var row transactionModel
	if err := db.Where("id = ?", id).First(&row).Error; err != nil {
		t.Fatalf("fetch updated row: %v", err)
	}
	if row.Status != "FAILED" {
		t.Fatalf("status = %q, want %q", row.Status, "FAILED")
	}
	if row.FailureReason == nil || *row.FailureReason != reason {
		t.Fatalf("failure_reason = %v, want %q", row.FailureReason, reason)
	}
}

// TestTransactionRepository_UpdateStatus_AlreadyTerminalIsNoOp proves a
// redelivered or out-of-order transaction.processed event for a
// transaction that already settled is a no-op, not an error — it must
// not re-open or overwrite a terminal transaction (docs/CLAUDE.md §3.5
// rule 3, §3.6 rule 4).
func TestTransactionRepository_UpdateStatus_AlreadyTerminalIsNoOp(t *testing.T) {
	db := newTestDB(t)
	repo := NewTransactionRepository(db)
	ctx := context.Background()

	id := mustNewID(t)
	if _, err := repo.Create(ctx, id, testRequest()); err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if err := repo.UpdateStatus(ctx, id, "COMPLETED", nil); err != nil {
		t.Fatalf("first UpdateStatus() error = %v, want nil", err)
	}

	// Redelivery of the same (or even a different) outcome must be a
	// silent no-op, never an error and never a re-write.
	if err := repo.UpdateStatus(ctx, id, "FAILED", nil); err != nil {
		t.Fatalf("redelivered UpdateStatus() error = %v, want nil (terminal transaction, must ack-and-drop)", err)
	}

	var row transactionModel
	if err := db.Where("id = ?", id).First(&row).Error; err != nil {
		t.Fatalf("fetch row: %v", err)
	}
	if row.Status != "COMPLETED" {
		t.Fatalf("status = %q after redelivered event, want unchanged %q", row.Status, "COMPLETED")
	}
}

func TestTransactionRepository_UpdateStatus_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewTransactionRepository(db)

	err := repo.UpdateStatus(context.Background(), mustNewID(t), "COMPLETED", nil)
	if !errors.Is(err, domain.ErrTransactionNotFound) {
		t.Fatalf("UpdateStatus() error = %v, want %v", err, domain.ErrTransactionNotFound)
	}
}

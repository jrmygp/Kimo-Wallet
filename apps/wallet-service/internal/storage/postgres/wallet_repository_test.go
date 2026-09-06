package postgres

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/domain"
	"github.com/jrmygp/kimo-wallet/apps/wallet-service/migrations"
)

// newTestDB connects to a real Postgres instance for integration testing.
// Skips (does not fail) when WALLET_SERVICE_TEST_DATABASE_URL is unset —
// same convention as user-service's newTestDB.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	databaseURL := os.Getenv("WALLET_SERVICE_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("WALLET_SERVICE_TEST_DATABASE_URL not set; skipping Postgres integration test")
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

	if err := db.Exec("TRUNCATE TABLE wallets").Error; err != nil {
		t.Fatalf("truncate wallets table before test: %v", err)
	}

	return db
}

func TestWalletRepository_Create(t *testing.T) {
	db := newTestDB(t)
	repo := NewWalletRepository(db)
	ctx := context.Background()

	wallet, err := repo.Create(ctx, "11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222")
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if wallet.ID == "" || wallet.UserID != "22222222-2222-4222-8222-222222222222" {
		t.Fatalf("Create() returned unexpected wallet: %+v", wallet)
	}
	if wallet.Balance != 0 {
		t.Fatalf("Create() returned balance %d, want 0 (every wallet starts at zero)", wallet.Balance)
	}
	if wallet.Currency != "IDR" {
		t.Fatalf("Create() returned currency %q, want %q", wallet.Currency, "IDR")
	}
	if wallet.Status != "ACTIVE" {
		t.Fatalf("Create() returned status %q, want %q", wallet.Status, "ACTIVE")
	}
	if wallet.CreatedAt.IsZero() {
		t.Fatalf("Create() returned zero CreatedAt")
	}
}

// TestWalletRepository_Create_DuplicateUserID proves the one-wallet-per-user
// invariant is enforced by the database, and surfaces as
// domain.ErrWalletAlreadyExists — the signal WalletService.CreateWallet
// turns into an idempotent no-op rather than a failure.
func TestWalletRepository_Create_DuplicateUserID(t *testing.T) {
	db := newTestDB(t)
	repo := NewWalletRepository(db)
	ctx := context.Background()

	const userID = "22222222-2222-4222-8222-222222222222"

	if _, err := repo.Create(ctx, "11111111-1111-4111-8111-111111111111", userID); err != nil {
		t.Fatalf("first Create() error = %v, want nil", err)
	}

	_, err := repo.Create(ctx, "33333333-3333-4333-8333-333333333333", userID)
	if !errors.Is(err, domain.ErrWalletAlreadyExists) {
		t.Fatalf("second Create() error = %v, want %v", err, domain.ErrWalletAlreadyExists)
	}

	var count int64
	if err := db.Model(&walletModel{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		t.Fatalf("count wallets: %v", err)
	}
	if count != 1 {
		t.Fatalf("got %d wallet rows for one user after a rejected duplicate Create(), want exactly 1", count)
	}
}

func TestWalletRepository_GetByUserID(t *testing.T) {
	db := newTestDB(t)
	repo := NewWalletRepository(db)
	ctx := context.Background()

	const userID = "22222222-2222-4222-8222-222222222222"

	created, err := repo.Create(ctx, "11111111-1111-4111-8111-111111111111", userID)
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	got, err := repo.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v, want nil", err)
	}
	if got != created {
		t.Fatalf("GetByUserID() = %+v, want %+v", got, created)
	}
}

func TestWalletRepository_GetByUserID_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewWalletRepository(db)

	_, err := repo.GetByUserID(context.Background(), "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, domain.ErrWalletNotFound) {
		t.Fatalf("GetByUserID() error = %v, want %v", err, domain.ErrWalletNotFound)
	}
}

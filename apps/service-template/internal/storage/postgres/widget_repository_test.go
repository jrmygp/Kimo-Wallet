package postgres

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/jrmygp/kimo-wallet/apps/service-template/internal/domain"
	"github.com/jrmygp/kimo-wallet/apps/service-template/migrations"
)

// newTestDB connects to a real Postgres instance for integration testing.
// Skips (does not fail) when SERVICE_TEMPLATE_TEST_DATABASE_URL is unset —
// same convention as every other service in this repo (docs/CLAUDE.md
// §5.5). Point this at an isolated *_test database, never a real one.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	databaseURL := os.Getenv("SERVICE_TEMPLATE_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("SERVICE_TEMPLATE_TEST_DATABASE_URL not set; skipping Postgres integration test")
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

	if err := db.Exec("TRUNCATE TABLE widgets").Error; err != nil {
		t.Fatalf("truncate widgets table before test: %v", err)
	}

	return db
}

func TestWidgetRepository_Create(t *testing.T) {
	db := newTestDB(t)
	repo := NewWidgetRepository(db)
	ctx := context.Background()

	input := domain.CreateWidgetInput{Name: "My Widget"}
	widget, err := repo.Create(ctx, "11111111-1111-4111-8111-111111111111", input)
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if widget.Name != "My Widget" {
		t.Fatalf("Create() returned name %q, want %q", widget.Name, "My Widget")
	}
	if widget.Status != statusActive {
		t.Fatalf("Create() returned status %q, want %q", widget.Status, statusActive)
	}
	if widget.CreatedAt.IsZero() {
		t.Fatalf("Create() returned zero CreatedAt")
	}
}

func TestWidgetRepository_GetByID(t *testing.T) {
	db := newTestDB(t)
	repo := NewWidgetRepository(db)
	ctx := context.Background()

	const id = "11111111-1111-4111-8111-111111111111"
	created, err := repo.Create(ctx, id, domain.CreateWidgetInput{Name: "My Widget"})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	got, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID() error = %v, want nil", err)
	}
	if got != created {
		t.Fatalf("GetByID() = %+v, want %+v", got, created)
	}
}

func TestWidgetRepository_GetByID_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewWidgetRepository(db)

	_, err := repo.GetByID(context.Background(), "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, domain.ErrWidgetNotFound) {
		t.Fatalf("GetByID() error = %v, want %v", err, domain.ErrWidgetNotFound)
	}
}

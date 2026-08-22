package postgres

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/domain"
	"github.com/jrmygp/kimo-wallet/apps/user-service/migrations"
)

// newTestDB connects to a real Postgres instance for integration testing.
// Skips (does not fail) when USER_SERVICE_TEST_DATABASE_URL is unset, since
// that means no database is available in this environment — see
// docker-compose.yml for how to run one locally.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	databaseURL := os.Getenv("USER_SERVICE_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("USER_SERVICE_TEST_DATABASE_URL not set; skipping Postgres integration test")
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

	if err := db.Exec("TRUNCATE TABLE users").Error; err != nil {
		t.Fatalf("truncate users table before test: %v", err)
	}

	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	input := domain.RegisterInput{PhoneNumber: "+6281234567890", FullName: "Jane Doe"}

	user, err := repo.Create(ctx, "11111111-1111-4111-8111-111111111111", input)
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if user.ID == "" || user.PhoneNumber != input.PhoneNumber || user.FullName != input.FullName {
		t.Fatalf("Create() returned unexpected user: %+v", user)
	}
	if user.CreatedAt.IsZero() {
		t.Fatalf("Create() returned zero CreatedAt")
	}
}

func TestUserRepository_Create_DuplicatePhoneNumber(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	input := domain.RegisterInput{PhoneNumber: "+6281234567890", FullName: "Jane Doe"}

	if _, err := repo.Create(ctx, "11111111-1111-4111-8111-111111111111", input); err != nil {
		t.Fatalf("first Create() error = %v, want nil", err)
	}

	_, err := repo.Create(ctx, "22222222-2222-4222-8222-222222222222", input)
	if !errors.Is(err, domain.ErrPhoneNumberTaken) {
		t.Fatalf("second Create() error = %v, want %v", err, domain.ErrPhoneNumberTaken)
	}
}

// TestUserRepository_Create_ConcurrentSamePhoneNumber proves the uniqueness
// guarantee holds under a race, not just under sequential calls — two
// goroutines register the same phone number at once; exactly one must
// succeed and the loser must see ErrPhoneNumberTaken, never a second row.
func TestUserRepository_Create_ConcurrentSamePhoneNumber(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	input := domain.RegisterInput{PhoneNumber: "+6281234567890", FullName: "Jane Doe"}
	ids := []string{
		"11111111-1111-4111-8111-111111111111",
		"22222222-2222-4222-8222-222222222222",
	}

	var wg sync.WaitGroup
	errs := make([]error, len(ids))
	for i, id := range ids {
		wg.Add(1)
		go func(i int, id string) {
			defer wg.Done()
			_, errs[i] = repo.Create(ctx, id, input)
		}(i, id)
	}
	wg.Wait()

	successes, conflicts := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrPhoneNumberTaken):
			conflicts++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if successes != 1 || conflicts != 1 {
		t.Fatalf("got %d successes and %d conflicts, want exactly 1 and 1", successes, conflicts)
	}

	var rowCount int64
	if err := db.Raw("SELECT COUNT(*) FROM users WHERE phone_number = ?", input.PhoneNumber).Scan(&rowCount).Error; err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("got %d rows for phone number, want 1", rowCount)
	}
}

func TestUserRepository_Login(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, "11111111-1111-4111-8111-111111111111",
		domain.RegisterInput{PhoneNumber: "+6281234567890", FullName: "Jane Doe"})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	user, err := repo.Login(ctx, domain.LoginInput{PhoneNumber: "+6281234567890"})
	if err != nil {
		t.Fatalf("Login() error = %v, want nil", err)
	}
	if user.ID != created.ID || user.PhoneNumber != created.PhoneNumber || user.FullName != created.FullName {
		t.Fatalf("Login() returned %+v, want %+v", user, created)
	}
}

func TestUserRepository_Login_UnknownPhoneNumber(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	_, err := repo.Login(ctx, domain.LoginInput{PhoneNumber: "+6281234567890"})
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("Login() error = %v, want %v", err, domain.ErrUserNotFound)
	}
}

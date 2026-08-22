package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/domain"
)

// pgUniqueViolation is the PostgreSQL error code for a unique-constraint violation.
const pgUniqueViolation = "23505"

const uniquePhoneNumberConstraint = "users_phone_number_key"

// userModel is the GORM row mapping for the users table. Kept separate from
// domain.User so the domain package stays free of ORM struct tags.
type userModel struct {
	ID          string `gorm:"primaryKey"`
	PhoneNumber string
	FullName    string
	CreatedAt   time.Time // populated by GORM on Create via its CreatedAt convention
}

func (userModel) TableName() string { return "users" }

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user row. The phone number uniqueness check is
// enforced by the database constraint, not by a prior SELECT — a
// check-then-insert would race under concurrent registrations with the same
// phone number.
func (r *UserRepository) Create(ctx context.Context, id string, input domain.RegisterInput) (domain.User, error) {
	row := userModel{
		ID:          id,
		PhoneNumber: input.PhoneNumber,
		FullName:    input.FullName,
	}

	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation && pgErr.ConstraintName == uniquePhoneNumberConstraint {
			return domain.User{}, domain.ErrPhoneNumberTaken
		}
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}

	return domain.User{
		ID:          row.ID,
		PhoneNumber: row.PhoneNumber,
		FullName:    row.FullName,
		CreatedAt:   row.CreatedAt,
	}, nil
}

func (r *UserRepository) Login(ctx context.Context, input domain.LoginInput) (domain.User, error) {
	var row userModel

	if err := r.db.WithContext(ctx).Where("phone_number = ?", input.PhoneNumber).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("find user by phone number: %w", err)
	}

	return domain.User{
		ID:          row.ID,
		PhoneNumber: row.PhoneNumber,
		FullName:    row.FullName,
		CreatedAt:   row.CreatedAt,
	}, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (domain.User, error) {
	var row userModel

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("find user by id: %w", err)
	}

	return domain.User{
		ID:          row.ID,
		PhoneNumber: row.PhoneNumber,
		FullName:    row.FullName,
		CreatedAt:   row.CreatedAt,
	}, nil
}

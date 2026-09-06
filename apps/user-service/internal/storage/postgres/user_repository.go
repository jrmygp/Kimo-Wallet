package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	userv1 "github.com/jrmygp/kimo-wallet/apps/user-service/gen/user/v1"
	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/domain"
	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/idgen"
)

// pgUniqueViolation is the PostgreSQL error code for a unique-constraint violation.
const pgUniqueViolation = "23505"

const uniquePhoneNumberConstraint = "users_phone_number_key"
const uniqueKimoIDConstraint = "users_kimo_id_key"

// userModel is the GORM row mapping for the users table. Kept separate from
// domain.User so the domain package stays free of ORM struct tags.
type userModel struct {
	ID             string `gorm:"primaryKey"`
	PhoneNumber    string
	FullName       string
	CreatedAt      time.Time // populated by GORM on Create via its CreatedAt convention
	ProfilePicture *string
	KimoID         string `gorm:"uniqueIndex"`
}

func (userModel) TableName() string { return "users" }

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user row and a "user.created" outbox event in the
// same database transaction (the transactional outbox pattern —
// docs/CLAUDE.md §3.4 rule 3 / §5.4) — so the event can never be silently
// lost even if the process dies immediately after this commits. A
// separate relay (internal/outbox) publishes it to Kafka afterward.
//
// The phone number uniqueness check is enforced by the database
// constraint, not by a prior SELECT — a check-then-insert would race
// under concurrent registrations with the same phone number. Same
// reasoning for kimoID: a collision is reported via ErrKimoIDCollision
// for the caller (Register) to retry with a freshly generated one, rather
// than this layer checking first and racing.
func (r *UserRepository) Create(ctx context.Context, id, kimoID string, input domain.RegisterInput) (domain.User, error) {
	var created domain.User

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row := userModel{
			ID:          id,
			PhoneNumber: input.PhoneNumber,
			FullName:    input.FullName,
			KimoID:      kimoID,
		}

		if err := tx.Create(&row).Error; err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
				switch pgErr.ConstraintName {
				case uniquePhoneNumberConstraint:
					return domain.ErrPhoneNumberTaken
				case uniqueKimoIDConstraint:
					return domain.ErrKimoIDCollision
				}
			}
			return fmt.Errorf("insert user: %w", err)
		}

		eventID, err := idgen.NewV4()
		if err != nil {
			return fmt.Errorf("generate outbox event id: %w", err)
		}

		payload, err := protojson.Marshal(&userv1.UserCreated{
			UserId:    row.ID,
			CreatedAt: timestamppb.New(row.CreatedAt),
		})
		if err != nil {
			return fmt.Errorf("marshal user.created payload: %w", err)
		}

		outboxRow := outboxEventModel{
			ID:        eventID,
			EventType: EventTypeUserCreated,
			Payload:   payload,
		}
		if err := tx.Create(&outboxRow).Error; err != nil {
			return fmt.Errorf("insert outbox event: %w", err)
		}

		created = domain.User{
			ID:             row.ID,
			PhoneNumber:    row.PhoneNumber,
			FullName:       row.FullName,
			CreatedAt:      row.CreatedAt,
			ProfilePicture: row.ProfilePicture,
			KimoID:         row.KimoID,
		}
		return nil
	})
	if err != nil {
		return domain.User{}, err
	}

	return created, nil
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
		ID:             row.ID,
		PhoneNumber:    row.PhoneNumber,
		FullName:       row.FullName,
		CreatedAt:      row.CreatedAt,
		ProfilePicture: row.ProfilePicture,
		KimoID:         row.KimoID,
	}, nil
}

// GetUserByID looks a user up by their KimoID — the public-facing
// identifier (see domain.User.KimoID) — not the internal `id` primary key
// despite the method's name, kept to match the UserService interface and
// the GetUserByID RPC it backs.
func (r *UserRepository) GetUserByID(ctx context.Context, kimoID string) (domain.User, error) {
	var row userModel

	if err := r.db.WithContext(ctx).Where("kimo_id = ?", kimoID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("find user by kimo id: %w", err)
	}

	return domain.User{
		ID:             row.ID,
		PhoneNumber:    row.PhoneNumber,
		FullName:       row.FullName,
		CreatedAt:      row.CreatedAt,
		ProfilePicture: row.ProfilePicture,
		KimoID:         row.KimoID,
	}, nil
}

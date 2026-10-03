package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/gorm"

	"github.com/jackc/pgx/v5/pgconn"
	transactionv1 "github.com/jrmygp/kimo-wallet/apps/transaction-service/gen/transaction/v1"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/domain"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/idgen"
)

const pgUniqueViolation = "23505"
const uniqueIdempotencyKeyConstraint = "transactions_idempotency_key_sender_user_id_key"

// statusPending is the status a transaction is created with. Not exported
// beyond this package yet — nothing here transitions it further; see
// docs/CLAUDE.md §3.5 for the full state machine this will eventually need.
const statusPending = "PENDING"
const statusProcessing = "PROCESSING"

type transactionModel struct {
	ID               string `gorm:"primaryKey"`
	IdempotencyKey   string `gorm:"uniqueIndex:idx_transactions_idempotency_key_sender"`
	SenderUserID     string `gorm:"uniqueIndex:idx_transactions_idempotency_key_sender"`
	SenderWalletID   string
	ReceiverUserID   string
	ReceiverWalletID string
	Currency         string
	Amount           int64
	Status           string
	FailureReason    *string
	CreatedAt        time.Time // populated by GORM on Create via its CreatedAt convention
	UpdatedAt        time.Time
	CompletedAt      *time.Time
}

func (transactionModel) TableName() string { return "transactions" }

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, id string, request domain.TransactionRequest) (domain.Transaction, error) {
	var created domain.Transaction

	row := transactionModel{
		ID:               id,
		IdempotencyKey:   request.IdempotencyKey,
		SenderUserID:     request.SenderUserID,
		SenderWalletID:   request.SenderWalletID,
		ReceiverUserID:   request.ReceiverUserID,
		ReceiverWalletID: request.ReceiverWalletID,
		Currency:         request.Currency,
		Amount:           request.Amount,
		Status:           statusPending,
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation && pgErr.ConstraintName == uniqueIdempotencyKeyConstraint {
				return domain.ErrIdempotencyConflict
			}
			return fmt.Errorf("insert transaction: %w", err)
		}

		eventID, err := idgen.NewV4()
		if err != nil {
			return fmt.Errorf("generate outbox event id: %w", err)
		}

		payload, err := protojson.Marshal(&transactionv1.TransactionCreated{
			TransactionId:    row.ID,
			SenderWalletId:   row.SenderWalletID,
			ReceiverWalletId: row.ReceiverWalletID,
			Amount:           row.Amount,
			Currency:         row.Currency,
		})
		if err != nil {
			return fmt.Errorf("marshal transaction.created payload: %w", err)
		}

		outboxRow := outboxEventModel{
			ID:        eventID,
			EventType: EventTypeTransactionCreated,
			Payload:   payload,
		}
		if err := tx.Create(&outboxRow).Error; err != nil {
			return fmt.Errorf("insert outbox event: %w", err)
		}

		created = domain.Transaction{
			ID:               row.ID,
			IdempotencyKey:   row.IdempotencyKey,
			SenderUserID:     row.SenderUserID,
			SenderWalletID:   row.SenderWalletID,
			ReceiverUserID:   row.ReceiverUserID,
			ReceiverWalletID: row.ReceiverWalletID,
			Currency:         row.Currency,
			Amount:           row.Amount,
			Status:           row.Status,
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
			CompletedAt:      row.CompletedAt,
		}
		return nil
	})
	if err != nil {
		return domain.Transaction{}, err
	}

	return created, nil
}

func (r *TransactionRepository) UpdateStatus(ctx context.Context, transactionID string, status string, failureReason *string) error {
	now := time.Now()

	result := r.db.WithContext(ctx).Model(&transactionModel{}).Where("id = ? AND status IN (?)", transactionID, []string{statusPending, statusProcessing}).Updates(map[string]any{
		"status":         status,
		"failure_reason": failureReason,
		"completed_at":   now,
		"updated_at":     now,
	})

	if result.Error != nil {
		return fmt.Errorf("update transaction status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		// The guard above didn't match. Either this id doesn't exist (a real
		// anomaly worth reporting — this service created every transaction
		// it will ever be asked to update), or the row exists but is already
		// terminal, which happens on every Kafka redelivery/out-of-order
		// delivery of transaction.processed (docs/CLAUDE.md §3.6 rule 4).
		// A terminal transaction is never re-processed (§3.5 rule 3), so
		// that case must be a no-op, not an error the caller retries or
		// dead-letters forever.
		var current transactionModel
		err := r.db.WithContext(ctx).Select("id").Where("id = ?", transactionID).First(&current).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrTransactionNotFound
			}
			return fmt.Errorf("look up transaction after no-op status update: %w", err)
		}
		return nil
	}

	return nil
}

package postgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/domain"
)

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
	row := transactionModel{
		ID:               id,
		IdempotencyKey:   request.IdempotencyKey,
		SenderUserID:     request.SenderUserID,
		SenderWalletID:   request.SenderWalletID,
		ReceiverUserID:   request.ReceiverUserID,
		ReceiverWalletID: request.ReceiverWalletID,
		Currency:         request.Currency,
		Amount:           request.Amount,
		Status:           "Pending",
	}

	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Transaction{}, fmt.Errorf("insert transaction: %w", err)
	}

	return domain.Transaction{
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
	}, nil
}

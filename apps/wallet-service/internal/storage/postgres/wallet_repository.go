package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/domain"
)

// pgUniqueViolation is the PostgreSQL error code for a unique-constraint violation.
const pgUniqueViolation = "23505"

const uniqueUserIDConstraint = "wallets_user_id_key"

// defaultCurrency is the only currency this service creates wallets in
// today. No currency selection exists anywhere upstream yet — see
// docs/agent-logs for the day this was written.
const defaultCurrency = "IDR"

const statusActive = "ACTIVE"

// walletModel is the GORM row mapping for the wallets table. Kept
// separate from domain.Wallet so the domain package stays free of ORM
// struct tags — same convention as user-service's userModel.
type walletModel struct {
	ID        string `gorm:"primaryKey"`
	UserID    string `gorm:"uniqueIndex"`
	Balance   int64
	Currency  string
	Status    string
	CreatedAt time.Time
}

func (walletModel) TableName() string { return "wallets" }

type WalletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

// Create inserts a new wallet row for userID, with a zero balance in
// defaultCurrency. The one-wallet-per-user rule is enforced by the
// database's unique constraint on user_id, not a prior SELECT — a
// check-then-insert would race if this were ever called concurrently for
// the same user (e.g. a redelivered Kafka event processed while the
// first delivery is still in flight). A collision is reported as
// domain.ErrWalletAlreadyExists, not an error the caller should retry —
// see internal/service's WalletService.CreateWallet for how it's turned
// into an idempotent no-op.
func (r *WalletRepository) Create(ctx context.Context, id, userID string) (domain.Wallet, error) {
	row := walletModel{
		ID:       id,
		UserID:   userID,
		Balance:  0,
		Currency: defaultCurrency,
		Status:   statusActive,
	}

	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation && pgErr.ConstraintName == uniqueUserIDConstraint {
			return domain.Wallet{}, domain.ErrWalletAlreadyExists
		}
		return domain.Wallet{}, fmt.Errorf("insert wallet: %w", err)
	}

	return domain.Wallet{
		ID:        row.ID,
		UserID:    row.UserID,
		Balance:   row.Balance,
		Currency:  row.Currency,
		Status:    row.Status,
		CreatedAt: row.CreatedAt,
	}, nil
}

// GetByUserID looks up the wallet belonging to userID.
func (r *WalletRepository) GetByUserID(ctx context.Context, userID string) (domain.Wallet, error) {
	var row walletModel

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Wallet{}, domain.ErrWalletNotFound
		}
		return domain.Wallet{}, fmt.Errorf("find wallet by user id: %w", err)
	}

	return domain.Wallet{
		ID:        row.ID,
		UserID:    row.UserID,
		Balance:   row.Balance,
		Currency:  row.Currency,
		Status:    row.Status,
		CreatedAt: row.CreatedAt,
	}, nil
}

package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/domain"
	"gorm.io/gorm"
)

type ledgerEntryModel struct {
	ID            string `gorm:"primaryKey"`
	TransactionID string
	WalletID      string
	Direction     string
	Amount        int64
	Currency      string
	CreatedAt     time.Time
}

func (ledgerEntryModel) TableName() string { return "ledger_entries" }

type LedgerEntriesRepository struct {
	db *gorm.DB
}

func NewLedgerEntriesRepository(db *gorm.DB) *LedgerEntriesRepository {
	return &LedgerEntriesRepository{db: db}
}

func (r *LedgerEntriesRepository) Create(ctx context.Context, entryRequest domain.LedgerEntry) (domain.LedgerEntry, error) {
	var created domain.LedgerEntry

	if err := r.db.Create(&created).Error; err != nil {
		return domain.LedgerEntry{}, fmt.Errorf("insert ledger entry: %w", err)
	}

	created = domain.LedgerEntry{
		ID:            created.ID,
		TransactionID: created.TransactionID,
		WalletID:      created.WalletID,
		Direction:     created.Direction,
		Amount:        created.Amount,
		Currency:      created.Currency,
		CreatedAt:     created.CreatedAt,
	}

	return created, nil
}

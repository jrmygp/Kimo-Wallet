package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	walletv1 "github.com/jrmygp/kimo-wallet/apps/wallet-service/gen/wallet/v1"
	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/domain"
	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/idgen"
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

// GetWalletByUserID looks up the wallet belonging to userID.
func (r *WalletRepository) GetWalletByUserID(ctx context.Context, userID string) (domain.Wallet, error) {
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

func (r *WalletRepository) ApplyTransfer(ctx context.Context, transactionID string, senderWalletID string, receiverWalletID string, amount int64) (status string, failureReason *string, err error) {
	txErr := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock both wallets in a deterministic order (ascending id) —
		// required to avoid deadlocking against a concurrent transfer
		// running the opposite direction (receiver -> sender), see
		// docs/CLAUDE.md §3.4 rule 2. Lock order is independent of which
		// one is sender vs receiver; senderWalletID != receiverWalletID
		// is already guaranteed upstream (domain.NewTransactionRequest on
		// transaction-service's side rejects a self-transfer), so no
		// tie-break case to worry about here.
		firstID, secondID := senderWalletID, receiverWalletID
		if secondID < firstID {
			firstID, secondID = secondID, firstID
		}

		var first, second walletModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", firstID).First(&first).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if firstID == senderWalletID {
					return domain.ErrSenderWalletNotFound
				}
				return domain.ErrReceiverWalletNotFound
			}
			return fmt.Errorf("lock wallet %s: %w", firstID, err)
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", secondID).First(&second).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if secondID == senderWalletID {
					return domain.ErrSenderWalletNotFound
				}
				return domain.ErrReceiverWalletNotFound
			}
			return fmt.Errorf("lock wallet %s: %w", secondID, err)
		}

		senderWallet, receiverWallet := first, second
		if firstID != senderWalletID {
			senderWallet, receiverWallet = second, first
		}

		// TODO: continue here, still inside this same tx (the locks above
		// are only held for the duration of this transaction) — an atomic
		// conditional debit on senderWallet (balance >= amount, checking
		// rows-affected for insufficient balance), credit receiverWallet,
		// insert both ledger_entries rows (DEBIT + CREDIT), and the
		// outbox event. See docs/plans/transaction-settlement.md Phase 2
		// step 3.
		_ = senderWallet
		_ = receiverWallet

		if senderWallet.Balance < amount {
			return domain.ErrInsufficientBalance
		}

		newSenderBalance := senderWallet.Balance - amount
		if err := tx.Model(&walletModel{}).Where("id = ?", senderWallet.ID).Update("balance", newSenderBalance).Error; err != nil {
			return fmt.Errorf("update sender wallet balance: %w", err)
		}

		newReceiverBalance := receiverWallet.Balance + amount
		if err := tx.Model(&walletModel{}).Where("id = ?", receiverWallet.ID).Update("balance", newReceiverBalance).Error; err != nil {
			return fmt.Errorf("update receiver wallet balance: %w", err)
		}

		// Create ledger entry for debit (sender)
		ledgerDebitID, err := idgen.NewV4()
		if err != nil {
			return fmt.Errorf("generate ledger entry debit id: %w", err)
		}

		ledgerEntryRow := ledgerEntryModel{
			ID:            ledgerDebitID,
			TransactionID: transactionID,
			WalletID:      senderWallet.ID,
			Direction:     "DEBIT",
			Amount:        amount,
			Currency:      senderWallet.Currency,
		}

		if err := tx.Create(&ledgerEntryRow).Error; err != nil {
			return fmt.Errorf("insert ledger entry debit: %w", err)
		}

		// Create ledger entry for credit (receiver)
		ledgerCreditID, err := idgen.NewV4()
		if err != nil {
			return fmt.Errorf("generate ledger entry credit id: %w", err)
		}

		ledgerCreditRow := ledgerEntryModel{
			ID:            ledgerCreditID,
			TransactionID: transactionID,
			WalletID:      receiverWallet.ID,
			Direction:     "CREDIT",
			Amount:        amount,
			Currency:      receiverWallet.Currency,
		}

		if err := tx.Create(&ledgerCreditRow).Error; err != nil {
			return fmt.Errorf("insert ledger entry credit: %w", err)
		}

		// Create outboxRow
		eventID, err := idgen.NewV4()
		if err != nil {
			return fmt.Errorf("generate outbox event id: %w", err)
		}

		payload, err := protojson.Marshal(&walletv1.TransactionProcessed{
			TransactionId: transactionID,
			Status:        "SUCCESS",
			FailureReason: nil,
		})

		outboxRow := outboxEventModel{
			ID:        eventID,
			EventType: EventTypeTransactionProcessed,
			Payload:   payload,
		}
		if err := tx.Create(&outboxRow).Error; err != nil {
			return fmt.Errorf("insert outbox event: %w", err)
		}

		return nil
	})

	if txErr != nil {
		switch {
		case errors.Is(txErr, domain.ErrSenderWalletNotFound):
			reason := "sender wallet not found"
			return "FAILED", &reason, domain.ErrSenderWalletNotFound
		case errors.Is(txErr, domain.ErrReceiverWalletNotFound):
			reason := "receiver wallet not found"
			return "FAILED", &reason, domain.ErrReceiverWalletNotFound
		default:
			return "", nil, fmt.Errorf("apply transfer: %w", txErr)
		}
	}

	return "SUCCESS", nil, nil
}

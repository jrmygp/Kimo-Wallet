package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidLedgerEntryID = errors.New("invalid ledger entry id format")
	ErrInvalidTransactionID = errors.New("invalid transaction id format")
	ErrInvalidWalletID      = errors.New("invalid wallet id format")
	ErrInvalidDirection     = errors.New("invalid direction, must be IN or OUT")
	ErrInvalidCurrency      = errors.New("invalid currency format")
)

var ledgerEntryIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type LedgerEntry struct {
	ID            string
	TransactionID string
	WalletID      string
	Direction     string
	Amount        int64
	Currency      string
	CreatedAt     time.Time
}

func NewLedgerEntry(id, transactionID, walletID, direction string, amount int64, currency string) (LedgerEntry, error) {
	id = strings.TrimSpace(id)
	if !ledgerEntryIDPattern.MatchString(id) {
		return LedgerEntry{}, ErrInvalidLedgerEntryID
	}

	transactionID = strings.TrimSpace(transactionID)
	if !ledgerEntryIDPattern.MatchString(transactionID) {
		return LedgerEntry{}, ErrInvalidTransactionID
	}

	walletID = strings.TrimSpace(walletID)
	if !ledgerEntryIDPattern.MatchString(walletID) {
		return LedgerEntry{}, ErrInvalidWalletID
	}

	direction = strings.TrimSpace(direction)
	if direction != "IN" && direction != "OUT" {
		return LedgerEntry{}, ErrInvalidDirection
	}

	currency = strings.TrimSpace(currency)
	if currency == "" {
		return LedgerEntry{}, ErrInvalidCurrency
	}

	return LedgerEntry{
		ID:            id,
		TransactionID: transactionID,
		WalletID:      walletID,
		Direction:     direction,
		Amount:        amount,
		Currency:      currency,
		CreatedAt:     time.Now(),
	}, nil
}

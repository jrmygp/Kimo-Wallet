// Package domain holds this service's core types, validation, and sentinel
// errors — kept free of gRPC/ORM/HTTP concerns so it can be tested and read
// without any of those (see docs/CLAUDE.md §5.4).
package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrSenderNotFound        = errors.New("sender user not found")
	ErrReceiverNotFound      = errors.New("receiver user not found")
	ErrIdempotencyConflict   = errors.New("idempotency key conflict")
	ErrTransactionNotFound   = errors.New("transaction not found")
	ErrInvalidIdempotencyKey = errors.New("idempotency key is required")
	ErrInvalidID             = errors.New("id must be a valid uuid")
	ErrSenderIsReceiver      = errors.New("sender and receiver must be different users")
	ErrInvalidAmount         = errors.New("amount must be greater than zero")
	ErrInvalidCurrency       = errors.New("currency is required")
)

// idPattern matches a v4 UUID as produced by internal/idgen.NewV4 — the
// shape user-service/wallet-service ids are always in.
var idPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type Transaction struct {
	ID               string
	IdempotencyKey   string
	SenderUserID     string
	SenderWalletID   string
	ReceiverUserID   string
	ReceiverWalletID string
	Currency         string
	Amount           int64
	Status           string
	Note             *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CompletedAt      *time.Time
}

type TransactionRequest struct {
	IdempotencyKey string
	SenderUserID   string
	// SenderWalletID / ReceiverWalletID are never supplied by a caller (the
	// CreateTransaction contract has no such fields). NewTransactionRequest
	// leaves them empty; TransactionService.CreateTransaction fills them from
	// wallet-service's answer for each user id, so the stored transaction can
	// only ever name the sender's and receiver's own wallets (§6).
	SenderWalletID   string
	ReceiverUserID   string
	ReceiverWalletID string
	Currency         string
	Amount           int64
	Note             *string
}

// NewTransactionRequest validates raw request fields and returns a
// TransactionRequest, or the first validation error encountered.
// Validation here is a shape/UX check — it is not the authority on
// whether the sender/receiver actually exist (that's
// TransactionService.CreateTransaction, via wallet-service) or on
// whether the idempotency key has already been used (that's the
// database's own unique constraint — see
// storage/postgres's TransactionRepository.Create).
func NewTransactionRequest(idempotencyKey, senderUserID, receiverUserID, currency string, amount int64, note *string) (TransactionRequest, error) {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if idempotencyKey == "" {
		return TransactionRequest{}, ErrInvalidIdempotencyKey
	}

	if !idPattern.MatchString(senderUserID) || !idPattern.MatchString(receiverUserID) {
		return TransactionRequest{}, ErrInvalidID
	}

	if senderUserID == receiverUserID {
		return TransactionRequest{}, ErrSenderIsReceiver
	}

	if amount <= 0 {
		return TransactionRequest{}, ErrInvalidAmount
	}

	currency = strings.TrimSpace(currency)
	if currency == "" {
		return TransactionRequest{}, ErrInvalidCurrency
	}

	return TransactionRequest{
		IdempotencyKey: idempotencyKey,
		SenderUserID:   senderUserID,
		ReceiverUserID: receiverUserID,
		Currency:       currency,
		Amount:         amount,
		Note:           note,
	}, nil
}

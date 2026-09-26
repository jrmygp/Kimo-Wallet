// Package domain holds this service's core types, validation, and sentinel
// errors — kept free of gRPC/ORM/HTTP concerns so it can be tested and read
// without any of those (see docs/CLAUDE.md §5.4).
//
// TEMPLATE NOTE: "Widget" stands in for this service's real entity. Rename
// the file, the type, every identifier below, and widgets_table in the
// migration to match — see apps/service-template/README.md for the full
// checklist.
package domain

import (
	"errors"
	"regexp"
	"time"
)

var (
	ErrWidgetNotFound = errors.New("widget not found")
	ErrInvalidID      = errors.New("id must be a valid uuid")
	ErrInvalidName    = errors.New("name must be between 1 and 100 characters")
)

// idPattern matches a v4 UUID as produced by internal/idgen.NewV4.
var idPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

const maxNameLength = 100

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
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CompletedAt      *time.Time
}

type TransactionRequest struct {
	IdempotencyKey   string
	SenderUserID     string
	SenderWalletID   string
	ReceiverUserID   string
	ReceiverWalletID string
	Currency         string
	Amount           int64
}

func NewTransactionRequest(idemPotencyKey string, senderUserID string, senderWalletID string, receiverUserID string, receiverWalletID string, currency string, amount int64) (TransactionRequest, error) {
	return TransactionRequest{
		IdempotencyKey:   idemPotencyKey,
		SenderUserID:     senderUserID,
		SenderWalletID:   senderWalletID,
		ReceiverUserID:   receiverUserID,
		ReceiverWalletID: receiverWalletID,
		Currency:         currency,
		Amount:           amount,
	}, nil
}

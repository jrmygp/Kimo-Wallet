package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	// ErrWalletAlreadyExists means a wallet for this user_id already
	// exists — the unique constraint on wallets.user_id, not a prior
	// SELECT, is what actually catches this (see storage/postgres's
	// WalletRepository.Create). Not a failure: this is exactly what makes
	// a redelivered user.created event a safe no-op instead of a second
	// wallet — see internal/service's WalletService.CreateWallet.
	ErrWalletAlreadyExists = errors.New("wallet already exists for this user")
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrInvalidUserID       = errors.New("invalid user id format")
)

var userIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Wallet is the identity record owned by the Wallet Service for a single
// user's balance. Balance is integer minor units (e.g. rupiah), never a
// float — docs/CLAUDE.md §3.3 rule 4.
type Wallet struct {
	ID        string
	UserID    string
	Balance   int64
	Currency  string
	Status    string
	CreatedAt time.Time
}

func NewGetWalletByUserIDInput(userID string) (string, error) {
	userID = strings.TrimSpace(userID)
	if !userIDPattern.MatchString(userID) {
		return "", ErrInvalidUserID
	}

	return userID, nil
}

// Package service orchestrates id generation and persistence for wallets.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/domain"
	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/idgen"
)

// WalletRepository persists Wallet records.
type WalletRepository interface {
	Create(ctx context.Context, id, userID string) (domain.Wallet, error)
	GetByUserID(ctx context.Context, userID string) (domain.Wallet, error)
}

type WalletService struct {
	repo WalletRepository
}

func NewWalletService(repo WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}

// CreateWallet provisions a new wallet for userID. Idempotent: if userID
// already has a wallet (domain.ErrWalletAlreadyExists, from the
// database's own unique constraint — see storage/postgres's
// WalletRepository.Create), that existing wallet is fetched and returned
// as success, not an error. This is what makes it safe to call twice for
// the same user — e.g. a Kafka user.created redelivery — without ever
// producing two wallets or the caller having to special-case "already
// exists" itself (docs/CLAUDE.md §3.6 rule 1).
func (s *WalletService) CreateWallet(ctx context.Context, userID string) (domain.Wallet, error) {
	id, err := idgen.NewV4()
	if err != nil {
		return domain.Wallet{}, fmt.Errorf("create wallet: %w", err)
	}

	wallet, err := s.repo.Create(ctx, id, userID)
	if err == nil {
		return wallet, nil
	}
	if errors.Is(err, domain.ErrWalletAlreadyExists) {
		return s.repo.GetByUserID(ctx, userID)
	}
	return domain.Wallet{}, err
}

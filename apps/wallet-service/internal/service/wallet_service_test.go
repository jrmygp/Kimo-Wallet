package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/domain"
)

type stubWalletRepository struct {
	createErr            error
	createdWallet        domain.Wallet
	existingWallet       domain.Wallet
	getWalletByUserIDErr error
	// createCalls records every userID Create was called with.
	createCalls []string
}

func (s *stubWalletRepository) Create(ctx context.Context, id, userID string) (domain.Wallet, error) {
	s.createCalls = append(s.createCalls, userID)
	if s.createErr != nil {
		return domain.Wallet{}, s.createErr
	}
	return s.createdWallet, nil
}

func (s *stubWalletRepository) GetWalletByUserID(ctx context.Context, userID string) (domain.Wallet, error) {
	if s.getWalletByUserIDErr != nil {
		return domain.Wallet{}, s.getWalletByUserIDErr
	}
	return s.existingWallet, nil
}

func TestWalletService_CreateWallet_Success(t *testing.T) {
	want := domain.Wallet{ID: "wallet-1", UserID: "user-1", Balance: 0, Currency: "IDR", Status: "ACTIVE"}
	repo := &stubWalletRepository{createdWallet: want}
	svc := NewWalletService(repo)

	got, err := svc.CreateWallet(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("CreateWallet() error = %v, want nil", err)
	}
	if got != want {
		t.Fatalf("CreateWallet() = %+v, want %+v", got, want)
	}
}

// TestWalletService_CreateWallet_AlreadyExists_IsIdempotentSuccess proves
// the whole point of the design: a second CreateWallet for the same user
// (e.g. a redelivered user.created event) returns the *existing* wallet
// as success, not an error the caller has to special-case.
func TestWalletService_CreateWallet_AlreadyExists_IsIdempotentSuccess(t *testing.T) {
	existing := domain.Wallet{ID: "wallet-1", UserID: "user-1", Balance: 5000, Currency: "IDR", Status: "ACTIVE"}
	repo := &stubWalletRepository{createErr: domain.ErrWalletAlreadyExists, existingWallet: existing}
	svc := NewWalletService(repo)

	got, err := svc.CreateWallet(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("CreateWallet() error = %v, want nil (already-exists must be idempotent success)", err)
	}
	if got != existing {
		t.Fatalf("CreateWallet() = %+v, want the existing wallet %+v", got, existing)
	}
}

func TestWalletService_CreateWallet_UnexpectedRepositoryError(t *testing.T) {
	repo := &stubWalletRepository{createErr: errors.New("connection refused")}
	svc := NewWalletService(repo)

	_, err := svc.CreateWallet(context.Background(), "user-1")
	if err == nil {
		t.Fatalf("CreateWallet() error = nil, want an error for an unexpected repository failure")
	}
	if errors.Is(err, domain.ErrWalletAlreadyExists) {
		t.Fatalf("CreateWallet() misreported an unexpected error as ErrWalletAlreadyExists")
	}
}

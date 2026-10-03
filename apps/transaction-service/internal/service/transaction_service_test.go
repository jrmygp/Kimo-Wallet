package service

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	walletv1 "github.com/jrmygp/kimo-wallet/apps/transaction-service/gen/wallet/v1"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/domain"
)

const (
	senderUserID     = "11111111-1111-4111-8111-111111111111"
	senderWalletID   = "22222222-2222-4222-8222-222222222222"
	receiverUserID   = "33333333-3333-4333-8333-333333333333"
	receiverWalletID = "44444444-4444-4444-8444-444444444444"
)

func validRequest() domain.TransactionRequest {
	return domain.TransactionRequest{
		IdempotencyKey:   "intent-123",
		SenderUserID:     senderUserID,
		SenderWalletID:   senderWalletID,
		ReceiverUserID:   receiverUserID,
		ReceiverWalletID: receiverWalletID,
		Currency:         "IDR",
		Amount:           5000,
	}
}

type stubTransactionRepository struct {
	createErr          error
	createdTransaction domain.Transaction
	createCalls        []domain.TransactionRequest

	updateStatusErr   error
	updateStatusCalls int
}

func (s *stubTransactionRepository) Create(ctx context.Context, id string, request domain.TransactionRequest) (domain.Transaction, error) {
	s.createCalls = append(s.createCalls, request)
	if s.createErr != nil {
		return domain.Transaction{}, s.createErr
	}
	return s.createdTransaction, nil
}

func (s *stubTransactionRepository) UpdateStatus(ctx context.Context, transactionID string, status string, failureReason *string) error {
	s.updateStatusCalls++
	return s.updateStatusErr
}

// stubWalletServiceClient reports notFoundUserIDs as codes.NotFound (the
// real client's behaviour for a missing wallet, per
// TransactionService.CreateTransaction's own comment) and err for every
// other call, regardless of which user id is asked about.
type stubWalletServiceClient struct {
	notFoundUserIDs map[string]bool
	err             error
}

func (s *stubWalletServiceClient) GetWalletByUserID(ctx context.Context, in *walletv1.GetWalletByUserIDRequest, opts ...grpc.CallOption) (*walletv1.GetWalletByUserIDResponse, error) {
	if s.notFoundUserIDs[in.GetUserId()] {
		return nil, status.Error(codes.NotFound, "wallet not found")
	}
	if s.err != nil {
		return nil, s.err
	}
	return &walletv1.GetWalletByUserIDResponse{}, nil
}

func TestTransactionService_CreateTransaction_Success(t *testing.T) {
	want := domain.Transaction{ID: "txn-1", Status: "PENDING"}
	repo := &stubTransactionRepository{createdTransaction: want}
	wallet := &stubWalletServiceClient{}
	svc := NewTransactionService(repo, wallet)

	got, err := svc.CreateTransaction(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("CreateTransaction() error = %v, want nil", err)
	}
	if got != want {
		t.Fatalf("CreateTransaction() = %+v, want %+v", got, want)
	}
	if len(repo.createCalls) != 1 {
		t.Fatalf("repo.Create called %d times, want 1", len(repo.createCalls))
	}
}

func TestTransactionService_CreateTransaction_SenderNotFound(t *testing.T) {
	repo := &stubTransactionRepository{}
	wallet := &stubWalletServiceClient{notFoundUserIDs: map[string]bool{senderUserID: true}}
	svc := NewTransactionService(repo, wallet)

	_, err := svc.CreateTransaction(context.Background(), validRequest())
	if !errors.Is(err, domain.ErrSenderNotFound) {
		t.Fatalf("CreateTransaction() error = %v, want %v", err, domain.ErrSenderNotFound)
	}
	if len(repo.createCalls) != 0 {
		t.Fatalf("repo.Create called %d times, want 0 — sender lookup failed before persistence", len(repo.createCalls))
	}
}

func TestTransactionService_CreateTransaction_ReceiverNotFound(t *testing.T) {
	repo := &stubTransactionRepository{}
	wallet := &stubWalletServiceClient{notFoundUserIDs: map[string]bool{receiverUserID: true}}
	svc := NewTransactionService(repo, wallet)

	_, err := svc.CreateTransaction(context.Background(), validRequest())
	if !errors.Is(err, domain.ErrReceiverNotFound) {
		t.Fatalf("CreateTransaction() error = %v, want %v", err, domain.ErrReceiverNotFound)
	}
	if len(repo.createCalls) != 0 {
		t.Fatalf("repo.Create called %d times, want 0 — receiver lookup failed before persistence", len(repo.createCalls))
	}
}

// TestTransactionService_CreateTransaction_WalletLookupInfrastructureError
// proves a genuine infrastructure failure (wallet-service unreachable) is
// never misreported as a business "not found" outcome.
func TestTransactionService_CreateTransaction_WalletLookupInfrastructureError(t *testing.T) {
	repo := &stubTransactionRepository{}
	wallet := &stubWalletServiceClient{err: errors.New("connection refused")}
	svc := NewTransactionService(repo, wallet)

	_, err := svc.CreateTransaction(context.Background(), validRequest())
	if err == nil {
		t.Fatalf("CreateTransaction() error = nil, want an error for an unreachable wallet-service")
	}
	if errors.Is(err, domain.ErrSenderNotFound) || errors.Is(err, domain.ErrReceiverNotFound) {
		t.Fatalf("CreateTransaction() misreported an infrastructure error as a not-found business outcome: %v", err)
	}
}

// TestTransactionService_CreateTransaction_ValidationFailsBeforeWalletLookup
// proves request validation happens first — an invalid request never
// reaches wallet-service at all.
func TestTransactionService_CreateTransaction_ValidationFailsBeforeWalletLookup(t *testing.T) {
	repo := &stubTransactionRepository{}
	wallet := &stubWalletServiceClient{err: errors.New("should never be called")}
	svc := NewTransactionService(repo, wallet)

	request := validRequest()
	request.ReceiverUserID = request.SenderUserID

	_, err := svc.CreateTransaction(context.Background(), request)
	if !errors.Is(err, domain.ErrSenderIsReceiver) {
		t.Fatalf("CreateTransaction() error = %v, want %v", err, domain.ErrSenderIsReceiver)
	}
}

func TestTransactionService_CreateTransaction_IdempotencyConflictPropagates(t *testing.T) {
	repo := &stubTransactionRepository{createErr: domain.ErrIdempotencyConflict}
	wallet := &stubWalletServiceClient{}
	svc := NewTransactionService(repo, wallet)

	_, err := svc.CreateTransaction(context.Background(), validRequest())
	if !errors.Is(err, domain.ErrIdempotencyConflict) {
		t.Fatalf("CreateTransaction() error = %v, want %v", err, domain.ErrIdempotencyConflict)
	}
}

func TestTransactionService_UpdateTransactionStatus_Success(t *testing.T) {
	repo := &stubTransactionRepository{}
	svc := NewTransactionService(repo, &stubWalletServiceClient{})

	if err := svc.UpdateTransactionStatus(context.Background(), "txn-1", "COMPLETED", nil); err != nil {
		t.Fatalf("UpdateTransactionStatus() error = %v, want nil", err)
	}
	if repo.updateStatusCalls != 1 {
		t.Fatalf("repo.UpdateStatus called %d times, want 1", repo.updateStatusCalls)
	}
}

func TestTransactionService_UpdateTransactionStatus_PropagatesRepositoryError(t *testing.T) {
	repo := &stubTransactionRepository{updateStatusErr: domain.ErrTransactionNotFound}
	svc := NewTransactionService(repo, &stubWalletServiceClient{})

	err := svc.UpdateTransactionStatus(context.Background(), "txn-1", "COMPLETED", nil)
	if !errors.Is(err, domain.ErrTransactionNotFound) {
		t.Fatalf("UpdateTransactionStatus() error = %v, want %v", err, domain.ErrTransactionNotFound)
	}
}

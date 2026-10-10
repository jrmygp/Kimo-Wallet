// Package service orchestrates id generation, input validation, and
// persistence for transactions — the layer grpcserver calls into and
// storage/postgres implements against.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/domain"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/idgen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	walletv1 "github.com/jrmygp/kimo-wallet/apps/transaction-service/gen/wallet/v1"
)

type TransactionRepository interface {
	Create(ctx context.Context, id string, request domain.TransactionRequest) (domain.Transaction, error)
	UpdateStatus(ctx context.Context, transactionID string, status string, failureReason *string) error
}

type WalletServiceClient interface {
	GetWalletByUserID(ctx context.Context, in *walletv1.GetWalletByUserIDRequest, opts ...grpc.CallOption) (*walletv1.GetWalletByUserIDResponse, error)
}

type TransactionService struct {
	repo   TransactionRepository
	wallet WalletServiceClient
}

func NewTransactionService(repo TransactionRepository, wallet WalletServiceClient) *TransactionService {
	return &TransactionService{repo: repo, wallet: wallet}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, request domain.TransactionRequest) (domain.Transaction, error) {
	validated, err := domain.NewTransactionRequest(
		request.IdempotencyKey,
		request.SenderUserID,
		request.ReceiverUserID,
		request.Currency,
		request.Amount,
	)
	if err != nil {
		return domain.Transaction{}, err
	}
	request = validated

	// Both wallet ids come from wallet-service's answer for each user id —
	// never from the caller — so the stored transaction (and the
	// transaction.created event wallet-service settles from) can only name
	// the sender's and receiver's own wallets (§6). GetWalletByUserID
	// returns an error (codes.NotFound) on a missing wallet — it never
	// succeeds with a nil Wallet — so NotFound must be read off the error
	// itself, not the response.
	senderWalletID, err := s.lookupWalletID(ctx, request.SenderUserID, domain.ErrSenderNotFound)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("sender wallet lookup: %w", err)
	}
	receiverWalletID, err := s.lookupWalletID(ctx, request.ReceiverUserID, domain.ErrReceiverNotFound)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("receiver wallet lookup: %w", err)
	}
	request.SenderWalletID = senderWalletID
	request.ReceiverWalletID = receiverWalletID

	id, err := idgen.NewV4()
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("create transaction: %w", err)
	}

	return s.repo.Create(ctx, id, request)
}

// lookupWalletID returns userID's wallet id, notFound when wallet-service
// reports codes.NotFound, or a wrapped error for anything else (including a
// successful response with no wallet id, which would otherwise be persisted
// as an empty wallet reference).
func (s *TransactionService) lookupWalletID(ctx context.Context, userID string, notFound error) (string, error) {
	resp, err := s.wallet.GetWalletByUserID(ctx, &walletv1.GetWalletByUserIDRequest{UserId: userID})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return "", notFound
		}
		return "", err
	}
	walletID := resp.GetWallet().GetId()
	if walletID == "" {
		return "", errors.New("wallet-service returned a wallet with no id")
	}
	return walletID, nil
}

func (s *TransactionService) UpdateTransactionStatus(ctx context.Context, transactionID string, status string, failureReason *string) error {
	if err := s.repo.UpdateStatus(ctx, transactionID, status, failureReason); err != nil {
		return fmt.Errorf("update transaction status: %w", err)
	}
	return nil
}

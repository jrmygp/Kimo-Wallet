// Package service orchestrates id generation, input validation, and
// persistence for widgets — the layer grpcserver calls into and
// storage/postgres implements against.
package service

import (
	"context"
	"fmt"

	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/domain"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/idgen"
	"google.golang.org/grpc"

	walletv1 "github.com/jrmygp/kimo-wallet/apps/transaction-service/gen/wallet/v1"
)

type TransactionRepository interface {
	Create(ctx context.Context, id string, request domain.TransactionRequest) (domain.Transaction, error)
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
	if senderWallet, err := s.wallet.GetWalletByUserID(ctx, &walletv1.GetWalletByUserIDRequest{UserId: request.SenderUserID}); err != nil {
		return domain.Transaction{}, fmt.Errorf("sender wallet lookup: %w", err)
	} else if senderWallet.Wallet == nil {
		return domain.Transaction{}, domain.ErrSenderNotFound
	}

	if receiverWallet, err := s.wallet.GetWalletByUserID(ctx, &walletv1.GetWalletByUserIDRequest{UserId: request.ReceiverUserID}); err != nil {
		return domain.Transaction{}, fmt.Errorf("receiver wallet lookup: %w", err)
	} else if receiverWallet.Wallet == nil {
		return domain.Transaction{}, domain.ErrReceiverNotFound
	}

	id, err := idgen.NewV4()
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("create transaction: %w", err)
	}

	return s.repo.Create(ctx, id, request)
}

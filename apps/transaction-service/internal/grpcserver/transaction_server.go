package grpcserver

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jrmygp/kimo-wallet/apps/service-template/internal/domain"
	transactionv1 "github.com/jrmygp/kimo-wallet/apps/transaction-service/gen/transaction/v1"
)

type transactionService interface {
	CreateTransaction(ctx context.Context, userID string, walletID string, amount int64, kimoID string) (domain.Transaction, error)
}

func toProtoTransaction(transaction domain.Transaction) *transactionv1.Transaction {
	return &transactionv1.Transaction{
		Id:        transaction.ID,
		UserID:    transaction.UserID,
		WalletID:  transaction.WalletID,
		Amount:    transaction.Amount,
		CreatedAt: timestamppb.New(transaction.CreatedAt),
	}
}

type TransactionServer struct {
	transactionv1.UnimplementedWidgetServiceServer
	service transactionService
}

func NewWidgetServer(service transactionService) *TransactionServer {
	return &TransactionServer{service: service}
}

func (s *TransactionServer) CreateTransaction(ctx context.Context, req *transactionv1.CreateWidgetRequest) (*transactionv1.CreateTransactionResponse, error) {
	transaction, err := s.service.CreateTransaction(ctx, req.GetName())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidName):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to create transaction")
		}
	}

	return &transactionv1.CreateTransactionResponse{Transaction: toProtoTransaction(transaction)}, nil
}

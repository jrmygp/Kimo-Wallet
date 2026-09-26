package grpcserver

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	transactionv1 "github.com/jrmygp/kimo-wallet/apps/transaction-service/gen/transaction/v1"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/domain"
)

type transactionService interface {
	CreateTransaction(ctx context.Context, request domain.TransactionRequest) (domain.Transaction, error)
}

func toProtoTransaction(transaction domain.Transaction) *transactionv1.Transaction {
	return &transactionv1.Transaction{
		Id:               transaction.ID,
		IdempotencyKey:   transaction.IdempotencyKey,
		SenderUserId:     transaction.SenderUserID,
		SenderWalletId:   transaction.SenderWalletID,
		ReceiverUserId:   transaction.ReceiverUserID,
		ReceiverWalletId: transaction.ReceiverWalletID,
		Amount:           transaction.Amount,
		Currency:         transaction.Currency,
		Status:           transaction.Status,
		CreatedAt:        timestamppb.New(transaction.CreatedAt),
	}
}

type TransactionServer struct {
	transactionv1.UnimplementedTransactionServiceServer
	service transactionService
}

func NewTransactionServer(service transactionService) *TransactionServer {
	return &TransactionServer{service: service}
}

func (s *TransactionServer) CreateTransaction(ctx context.Context, req *transactionv1.CreateTransactionRequest) (*transactionv1.CreateTransactionResponse, error) {
	transaction, err := s.service.CreateTransaction(ctx, domain.TransactionRequest{
		IdempotencyKey:   req.IdempotencyKey,
		SenderUserID:     req.SenderUserId,
		SenderWalletID:   req.SenderWalletId,
		ReceiverUserID:   req.ReceiverUserId,
		ReceiverWalletID: req.ReceiverWalletId,
		Currency:         req.Currency,
		Amount:           req.Amount,
	})
	if err != nil {
		// TODO: no transaction-specific sentinel errors exist yet (no
		// domain-level amount/sender/receiver validation, no typed
		// idempotency-conflict error from the repository's unique
		// constraint) — everything maps to Internal for now. Once those
		// exist, switch on them here the same way wallet_server.go does
		// (e.g. codes.InvalidArgument for a bad amount, codes.AlreadyExists
		// or similar for a replayed idempotency key with a different body —
		// see docs/CLAUDE.md §3.2 rule 5).
		return nil, status.Error(codes.Internal, "failed to create transaction")
	}

	return &transactionv1.CreateTransactionResponse{Transaction: toProtoTransaction(transaction)}, nil
}

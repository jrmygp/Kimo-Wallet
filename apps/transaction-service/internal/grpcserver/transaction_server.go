package grpcserver

import (
	"context"
	"errors"

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
		switch {
		case errors.Is(err, domain.ErrInvalidIdempotencyKey),
			errors.Is(err, domain.ErrInvalidID),
			errors.Is(err, domain.ErrSenderIsReceiver),
			errors.Is(err, domain.ErrInvalidAmount),
			errors.Is(err, domain.ErrInvalidCurrency):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, domain.ErrSenderNotFound), errors.Is(err, domain.ErrReceiverNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, domain.ErrIdempotencyConflict):
			// A repeated idempotency key should really return the
			// *original* transaction's result, not an error at all (see
			// docs/CLAUDE.md §3.2 rule 4) — that lookup doesn't exist yet
			// (no TransactionRepository.GetByIdempotencyKey), so this is a
			// conservative placeholder: better to reject the retry loudly
			// than to silently create a duplicate.
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to create transaction")
		}
	}

	return &transactionv1.CreateTransactionResponse{Transaction: toProtoTransaction(transaction)}, nil
}

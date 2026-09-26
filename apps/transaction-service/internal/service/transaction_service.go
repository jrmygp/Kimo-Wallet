// Package service orchestrates id generation, input validation, and
// persistence for widgets — the layer grpcserver calls into and
// storage/postgres implements against.
package service

import (
	"context"
	"fmt"

	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/domain"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/idgen"
)

type TransactionRepository interface {
	Create(ctx context.Context, id string, request domain.TransactionRequest) (domain.Transaction, error)
}

type TransactionService struct {
	repo TransactionRepository
}

func NewTransactionService(repo TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, request domain.TransactionRequest) (domain.Transaction, error) {
	id, err := idgen.NewV4()
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("create transaction: %w", err)
	}

	return s.repo.Create(ctx, id, request)
}

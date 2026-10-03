package eventconsumer

import (
	"context"
	"errors"
	"fmt"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/encoding/protojson"

	walletv1 "github.com/jrmygp/kimo-wallet/apps/transaction-service/gen/wallet/v1"
)

type TransferProcessor interface {
	UpdateTransactionStatus(ctx context.Context, transactionID string, status string, failureReason *string) error
}

type TransactionProcessedProcessor struct {
	transfer TransferProcessor
}

func NewTransactionProcessedProcessor(transfer TransferProcessor) *TransactionProcessedProcessor {
	return &TransactionProcessedProcessor{transfer: transfer}
}

func (p *TransactionProcessedProcessor) Process(ctx context.Context, msg kafka.Message) error {
	var event walletv1.TransactionProcessed
	if err := protojson.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal transaction.processed payload: %w", err)
	}
	if event.GetTransactionId() == "" {
		return errors.New("transaction.processed event missing transaction_id")
	}
	if event.GetStatus() == "" {
		return errors.New("transaction.processed event missing status")
	}

	if err := p.transfer.UpdateTransactionStatus(ctx, event.GetTransactionId(), event.GetStatus(), event.FailureReason); err != nil {
		return fmt.Errorf("update transaction status for transaction %s: %w", event.GetTransactionId(), err)
	}
	return nil
}

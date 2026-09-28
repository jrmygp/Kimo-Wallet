package eventconsumer

import (
	"context"
	"errors"
	"fmt"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/encoding/protojson"

	transactionv1 "github.com/jrmygp/kimo-wallet/apps/wallet-service/gen/transaction/v1"
)

// TransferApplier is the subset of service.WalletService this processor
// needs.
type TransferApplier interface {
	ApplyTransfer(ctx context.Context, transactionID, senderWalletID, receiverWalletID string, amount int64, currency string) (status string, failureReason *string, err error)
}

// TransactionCreatedProcessor handles "transaction.created" events:
// debiting the sender, crediting the receiver, and recording the ledger
// and outcome — see storage/postgres's WalletRepository.ApplyTransfer and
// docs/plans/transaction-settlement.md. See Handler for the retry/DLQ
// machinery this runs under.
type TransactionCreatedProcessor struct {
	wallets TransferApplier
}

func NewTransactionCreatedProcessor(wallets TransferApplier) *TransactionCreatedProcessor {
	return &TransactionCreatedProcessor{wallets: wallets}
}

func (p *TransactionCreatedProcessor) Process(ctx context.Context, msg kafka.Message) error {
	var event transactionv1.TransactionCreated
	if err := protojson.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal transaction.created payload: %w", err)
	}
	if event.GetTransactionId() == "" {
		return errors.New("transaction.created event missing transaction_id")
	}
	if event.GetSenderWalletId() == "" || event.GetReceiverWalletId() == "" {
		return errors.New("transaction.created event missing sender_wallet_id or receiver_wallet_id")
	}

	// ApplyTransfer's own err is only ever non-nil for a genuine
	// infrastructure failure — "sender/receiver not found" and
	// "insufficient balance" are business outcomes it already records and
	// reports itself (via its own transaction.processed outbox event),
	// not something this processor needs to inspect further. A non-nil
	// err here correctly triggers Handler's normal bounded-retry/DLQ path.
	if _, _, err := p.wallets.ApplyTransfer(ctx, event.GetTransactionId(), event.GetSenderWalletId(), event.GetReceiverWalletId(), event.GetAmount(), event.GetCurrency()); err != nil {
		return fmt.Errorf("apply transfer for transaction %s: %w", event.GetTransactionId(), err)
	}
	return nil
}

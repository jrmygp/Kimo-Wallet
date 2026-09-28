package eventconsumer

import (
	"context"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/encoding/protojson"

	transactionv1 "github.com/jrmygp/kimo-wallet/apps/wallet-service/gen/transaction/v1"
)

type stubTransferApplier struct {
	status    string
	reason    *string
	err       error
	callCount int
	// lastArgs records what ApplyTransfer was last called with, so tests
	// can prove the event's fields actually made it through unmarshal.
	lastTransactionID    string
	lastSenderWalletID   string
	lastReceiverWalletID string
	lastAmount           int64
	lastCurrency         string
}

func (s *stubTransferApplier) ApplyTransfer(ctx context.Context, transactionID, senderWalletID, receiverWalletID string, amount int64, currency string) (string, *string, error) {
	s.callCount++
	s.lastTransactionID = transactionID
	s.lastSenderWalletID = senderWalletID
	s.lastReceiverWalletID = receiverWalletID
	s.lastAmount = amount
	s.lastCurrency = currency
	return s.status, s.reason, s.err
}

func mustMarshalTransactionCreated(t *testing.T, event *transactionv1.TransactionCreated) []byte {
	t.Helper()
	payload, err := protojson.Marshal(event)
	if err != nil {
		t.Fatalf("marshal test payload: %v", err)
	}
	return payload
}

func TestTransactionCreatedProcessor_Success(t *testing.T) {
	wallets := &stubTransferApplier{status: "COMPLETED"}
	p := NewTransactionCreatedProcessor(wallets)

	msg := kafka.Message{Value: mustMarshalTransactionCreated(t, &transactionv1.TransactionCreated{
		TransactionId:    "txn-1",
		SenderWalletId:   "sender-1",
		ReceiverWalletId: "receiver-1",
		Amount:           5000,
		Currency:         "IDR",
	})}

	if err := p.Process(context.Background(), msg); err != nil {
		t.Fatalf("Process() error = %v, want nil", err)
	}
	if wallets.callCount != 1 {
		t.Fatalf("ApplyTransfer called %d times, want 1", wallets.callCount)
	}
	if wallets.lastTransactionID != "txn-1" || wallets.lastSenderWalletID != "sender-1" ||
		wallets.lastReceiverWalletID != "receiver-1" || wallets.lastAmount != 5000 || wallets.lastCurrency != "IDR" {
		t.Fatalf("ApplyTransfer called with unexpected args: %+v", wallets)
	}
}

// TestTransactionCreatedProcessor_FailedOutcomeIsNotAnError proves that
// ApplyTransfer reporting a business failure (sender/receiver not found,
// insufficient balance — status "FAILED", err nil) is NOT treated as a
// processing error: Process must return nil, since ApplyTransfer already
// recorded and will publish the outcome itself. Only a genuine
// ApplyTransfer error (infrastructure failure) should make Process
// return an error and trigger Handler's retry/DLQ path.
func TestTransactionCreatedProcessor_FailedOutcomeIsNotAnError(t *testing.T) {
	reason := "insufficient balance"
	wallets := &stubTransferApplier{status: "FAILED", reason: &reason}
	p := NewTransactionCreatedProcessor(wallets)

	msg := kafka.Message{Value: mustMarshalTransactionCreated(t, &transactionv1.TransactionCreated{
		TransactionId:    "txn-1",
		SenderWalletId:   "sender-1",
		ReceiverWalletId: "receiver-1",
		Amount:           999999,
		Currency:         "IDR",
	})}

	if err := p.Process(context.Background(), msg); err != nil {
		t.Fatalf("Process() error = %v, want nil — a FAILED outcome is not a processing error", err)
	}
}

func TestTransactionCreatedProcessor_PropagatesInfrastructureError(t *testing.T) {
	wallets := &stubTransferApplier{err: errors.New("connection refused")}
	p := NewTransactionCreatedProcessor(wallets)

	msg := kafka.Message{Value: mustMarshalTransactionCreated(t, &transactionv1.TransactionCreated{
		TransactionId:    "txn-1",
		SenderWalletId:   "sender-1",
		ReceiverWalletId: "receiver-1",
		Amount:           5000,
		Currency:         "IDR",
	})}

	if err := p.Process(context.Background(), msg); err == nil {
		t.Fatalf("Process() error = nil, want an error for a genuine ApplyTransfer failure")
	}
}

func TestTransactionCreatedProcessor_RejectsMalformedPayload(t *testing.T) {
	wallets := &stubTransferApplier{}
	p := NewTransactionCreatedProcessor(wallets)

	err := p.Process(context.Background(), kafka.Message{Value: []byte("not json at all")})
	if err == nil {
		t.Fatalf("Process() error = nil, want an error for a malformed payload")
	}
	if wallets.callCount != 0 {
		t.Fatalf("ApplyTransfer was called %d times for a malformed payload, want 0", wallets.callCount)
	}
}

func TestTransactionCreatedProcessor_RejectsMissingFields(t *testing.T) {
	wallets := &stubTransferApplier{}
	p := NewTransactionCreatedProcessor(wallets)

	tests := []struct {
		name  string
		event *transactionv1.TransactionCreated
	}{
		{"missing transaction_id", &transactionv1.TransactionCreated{SenderWalletId: "sender-1", ReceiverWalletId: "receiver-1", Amount: 5000, Currency: "IDR"}},
		{"missing sender_wallet_id", &transactionv1.TransactionCreated{TransactionId: "txn-1", ReceiverWalletId: "receiver-1", Amount: 5000, Currency: "IDR"}},
		{"missing receiver_wallet_id", &transactionv1.TransactionCreated{TransactionId: "txn-1", SenderWalletId: "sender-1", Amount: 5000, Currency: "IDR"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.Process(context.Background(), kafka.Message{Value: mustMarshalTransactionCreated(t, tt.event)})
			if err == nil {
				t.Fatalf("Process() error = nil, want an error for %s", tt.name)
			}
			if wallets.callCount != 0 {
				t.Fatalf("ApplyTransfer was called %d times for %s, want 0", wallets.callCount, tt.name)
			}
		})
	}
}

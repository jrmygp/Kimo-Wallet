package eventconsumer

import (
	"context"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/encoding/protojson"

	walletv1 "github.com/jrmygp/kimo-wallet/apps/transaction-service/gen/wallet/v1"
)

type stubTransferProcessor struct {
	err       error
	callCount int
	// lastArgs records what UpdateTransactionStatus was last called with,
	// so tests can prove the event's fields actually made it through
	// unmarshal.
	lastTransactionID string
	lastStatus        string
	lastFailureReason *string
}

func (s *stubTransferProcessor) UpdateTransactionStatus(ctx context.Context, transactionID string, status string, failureReason *string) error {
	s.callCount++
	s.lastTransactionID = transactionID
	s.lastStatus = status
	s.lastFailureReason = failureReason
	return s.err
}

func mustMarshalTransactionProcessed(t *testing.T, event *walletv1.TransactionProcessed) []byte {
	t.Helper()
	payload, err := protojson.Marshal(event)
	if err != nil {
		t.Fatalf("marshal test payload: %v", err)
	}
	return payload
}

func TestTransactionProcessedProcessor_Success(t *testing.T) {
	transfer := &stubTransferProcessor{}
	p := NewTransactionProcessedProcessor(transfer)

	msg := kafka.Message{Value: mustMarshalTransactionProcessed(t, &walletv1.TransactionProcessed{
		TransactionId: "txn-1",
		Status:        "COMPLETED",
	})}

	if err := p.Process(context.Background(), msg); err != nil {
		t.Fatalf("Process() error = %v, want nil", err)
	}
	if transfer.callCount != 1 {
		t.Fatalf("UpdateTransactionStatus called %d times, want 1", transfer.callCount)
	}
	if transfer.lastTransactionID != "txn-1" || transfer.lastStatus != "COMPLETED" || transfer.lastFailureReason != nil {
		t.Fatalf("UpdateTransactionStatus called with unexpected args: %+v", transfer)
	}
}

func TestTransactionProcessedProcessor_FailedWithReason(t *testing.T) {
	transfer := &stubTransferProcessor{}
	p := NewTransactionProcessedProcessor(transfer)

	reason := "insufficient balance"
	msg := kafka.Message{Value: mustMarshalTransactionProcessed(t, &walletv1.TransactionProcessed{
		TransactionId: "txn-1",
		Status:        "FAILED",
		FailureReason: &reason,
	})}

	if err := p.Process(context.Background(), msg); err != nil {
		t.Fatalf("Process() error = %v, want nil", err)
	}
	if transfer.lastFailureReason == nil || *transfer.lastFailureReason != reason {
		t.Fatalf("UpdateTransactionStatus failureReason = %v, want %q", transfer.lastFailureReason, reason)
	}
}

// TestTransactionProcessedProcessor_RedeliveryIsNotAnError proves that a
// redelivered/out-of-order event for an already-terminal transaction —
// which UpdateTransactionStatus reports as success (nil error), per
// TransactionRepository.UpdateStatus's no-op guard — is never treated as
// a processing failure here. Process has no way to distinguish "really
// updated" from "already terminal, no-op" and must not need to: both are
// nil from UpdateTransactionStatus, and both must leave the message
// acked, never retried or dead-lettered (docs/CLAUDE.md §3.5 rule 3, §3.6
// rule 4).
func TestTransactionProcessedProcessor_RedeliveryIsNotAnError(t *testing.T) {
	transfer := &stubTransferProcessor{err: nil}
	p := NewTransactionProcessedProcessor(transfer)

	msg := kafka.Message{Value: mustMarshalTransactionProcessed(t, &walletv1.TransactionProcessed{
		TransactionId: "txn-1",
		Status:        "COMPLETED",
	})}

	if err := p.Process(context.Background(), msg); err != nil {
		t.Fatalf("Process() error = %v, want nil for a redelivered/no-op outcome", err)
	}
}

func TestTransactionProcessedProcessor_PropagatesInfrastructureError(t *testing.T) {
	transfer := &stubTransferProcessor{err: errors.New("connection refused")}
	p := NewTransactionProcessedProcessor(transfer)

	msg := kafka.Message{Value: mustMarshalTransactionProcessed(t, &walletv1.TransactionProcessed{
		TransactionId: "txn-1",
		Status:        "COMPLETED",
	})}

	if err := p.Process(context.Background(), msg); err == nil {
		t.Fatalf("Process() error = nil, want an error for a genuine UpdateTransactionStatus failure")
	}
}

func TestTransactionProcessedProcessor_RejectsMalformedPayload(t *testing.T) {
	transfer := &stubTransferProcessor{}
	p := NewTransactionProcessedProcessor(transfer)

	err := p.Process(context.Background(), kafka.Message{Value: []byte("not json at all")})
	if err == nil {
		t.Fatalf("Process() error = nil, want an error for a malformed payload")
	}
	if transfer.callCount != 0 {
		t.Fatalf("UpdateTransactionStatus was called %d times for a malformed payload, want 0", transfer.callCount)
	}
}

func TestTransactionProcessedProcessor_RejectsMissingTransactionID(t *testing.T) {
	transfer := &stubTransferProcessor{}
	p := NewTransactionProcessedProcessor(transfer)

	msg := kafka.Message{Value: mustMarshalTransactionProcessed(t, &walletv1.TransactionProcessed{
		Status: "COMPLETED",
	})}

	err := p.Process(context.Background(), msg)
	if err == nil {
		t.Fatalf("Process() error = nil, want an error for a missing transaction_id")
	}
	if transfer.callCount != 0 {
		t.Fatalf("UpdateTransactionStatus was called %d times for a missing transaction_id, want 0", transfer.callCount)
	}
}

func TestTransactionProcessedProcessor_RejectsInvalidStatus(t *testing.T) {
	transfer := &stubTransferProcessor{}
	p := NewTransactionProcessedProcessor(transfer)

	tests := []struct {
		name   string
		status string
	}{
		{"empty status", ""},
		{"unknown status", "PENDING"},
		{"lowercase variant", "completed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := kafka.Message{Value: mustMarshalTransactionProcessed(t, &walletv1.TransactionProcessed{
				TransactionId: "txn-1",
				Status:        tt.status,
			})}

			err := p.Process(context.Background(), msg)
			if err == nil {
				t.Fatalf("Process() error = nil, want an error for status %q", tt.status)
			}
			if transfer.callCount != 0 {
				t.Fatalf("UpdateTransactionStatus was called %d times for status %q, want 0", transfer.callCount, tt.status)
			}
		})
	}
}

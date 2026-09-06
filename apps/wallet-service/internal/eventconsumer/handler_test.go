package eventconsumer

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	userv1 "github.com/jrmygp/kimo-wallet/apps/wallet-service/gen/user/v1"
	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/domain"
)

type stubReader struct {
	committed []kafka.Message
}

func (s *stubReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	return kafka.Message{}, errors.New("not used in these tests — processWithRetry is called directly")
}

func (s *stubReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	s.committed = append(s.committed, msgs...)
	return nil
}

type stubWalletService struct {
	err       error
	callCount int
	// failFirstN makes CreateWallet fail for the first N calls, then
	// succeed — used to test that a transient failure recovers on retry.
	failFirstN int
}

func (s *stubWalletService) CreateWallet(ctx context.Context, userID string) (domain.Wallet, error) {
	s.callCount++
	if s.callCount <= s.failFirstN {
		return domain.Wallet{}, errors.New("transient failure")
	}
	return domain.Wallet{}, s.err
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func mustMarshalUserCreated(t *testing.T, userID string) []byte {
	t.Helper()
	payload, err := protojson.Marshal(&userv1.UserCreated{UserId: userID, CreatedAt: timestamppb.Now()})
	if err != nil {
		t.Fatalf("marshal test payload: %v", err)
	}
	return payload
}

func TestHandler_ProcessWithRetry_SuccessCommitsOnce(t *testing.T) {
	reader := &stubReader{}
	wallets := &stubWalletService{}
	h := NewHandler(reader, wallets, testLogger())

	msg := kafka.Message{Offset: 1, Value: mustMarshalUserCreated(t, "user-1")}
	h.processWithRetry(context.Background(), msg)

	if wallets.callCount != 1 {
		t.Fatalf("CreateWallet called %d times, want 1", wallets.callCount)
	}
	if len(reader.committed) != 1 {
		t.Fatalf("committed %d messages, want 1", len(reader.committed))
	}
}

// TestHandler_ProcessWithRetry_RecoversAfterTransientFailure proves a
// message that fails once (or a few times) but eventually succeeds is
// retried, not immediately given up on.
func TestHandler_ProcessWithRetry_RecoversAfterTransientFailure(t *testing.T) {
	reader := &stubReader{}
	wallets := &stubWalletService{failFirstN: 2} // fails twice, succeeds on the 3rd attempt
	h := &Handler{reader: reader, wallets: wallets, logger: testLogger(), retryBackoff: 0}

	msg := kafka.Message{Offset: 1, Value: mustMarshalUserCreated(t, "user-1")}
	h.processWithRetry(context.Background(), msg)

	if wallets.callCount != 3 {
		t.Fatalf("CreateWallet called %d times, want 3 (2 failures + 1 success)", wallets.callCount)
	}
	if len(reader.committed) != 1 {
		t.Fatalf("committed %d messages, want 1 — a message that eventually succeeds must still be committed", len(reader.committed))
	}
}

// TestHandler_ProcessWithRetry_GivesUpAndCommitsAfterMaxAttempts proves a
// message that never succeeds is still committed after exhausting
// retries — a poison message must not block every message behind it
// forever, even without a real dead-letter queue (see this day's log entry).
func TestHandler_ProcessWithRetry_GivesUpAndCommitsAfterMaxAttempts(t *testing.T) {
	reader := &stubReader{}
	wallets := &stubWalletService{err: errors.New("permanent failure")}
	// retryBackoff: 0 — a real Handler still runs the real bounded-retry
	// loop maxProcessAttempts times, just without actually sleeping
	// between attempts (docs/CLAUDE.md §5.5: tests never rely on sleeps).
	h := &Handler{reader: reader, wallets: wallets, logger: testLogger(), retryBackoff: 0}

	msg := kafka.Message{Offset: 1, Value: mustMarshalUserCreated(t, "user-1")}
	h.processWithRetry(context.Background(), msg)

	if wallets.callCount != maxProcessAttempts {
		t.Fatalf("CreateWallet called %d times, want exactly %d (bounded, not infinite)", wallets.callCount, maxProcessAttempts)
	}
	if len(reader.committed) != 1 {
		t.Fatalf("committed %d messages, want 1 — a permanently-failing message must still be committed after giving up, or it blocks every message behind it", len(reader.committed))
	}
}

func TestHandler_Process_RejectsMalformedPayload(t *testing.T) {
	reader := &stubReader{}
	wallets := &stubWalletService{}
	h := &Handler{reader: reader, wallets: wallets, logger: testLogger(), retryBackoff: 0}

	err := h.process(context.Background(), kafka.Message{Value: []byte("not json at all")})
	if err == nil {
		t.Fatalf("process() error = nil, want an error for a malformed payload")
	}
	if wallets.callCount != 0 {
		t.Fatalf("CreateWallet was called %d times for a malformed payload, want 0", wallets.callCount)
	}
}

func TestHandler_Process_RejectsMissingUserID(t *testing.T) {
	reader := &stubReader{}
	wallets := &stubWalletService{}
	h := &Handler{reader: reader, wallets: wallets, logger: testLogger(), retryBackoff: 0}

	err := h.process(context.Background(), kafka.Message{Value: mustMarshalUserCreated(t, "")})
	if err == nil {
		t.Fatalf("process() error = nil, want an error for a missing user_id")
	}
	if wallets.callCount != 0 {
		t.Fatalf("CreateWallet was called %d times for an event with no user_id, want 0", wallets.callCount)
	}
}

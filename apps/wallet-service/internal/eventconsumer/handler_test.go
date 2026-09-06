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

type dlqPublication struct {
	topic   string
	key     []byte
	value   []byte
	headers map[string]string
}

type stubDeadLetterPublisher struct {
	err          error
	publications []dlqPublication
	// failFirstN makes Publish fail for the first N calls, then succeed —
	// mirrors stubWalletService.failFirstN, for the same reason: proving
	// a transient failure recovers on retry rather than being given up
	// on immediately.
	failFirstN int
}

func (s *stubDeadLetterPublisher) Publish(ctx context.Context, topic string, key, value []byte, headers map[string]string) error {
	s.publications = append(s.publications, dlqPublication{topic: topic, key: key, value: value, headers: headers})
	if len(s.publications) <= s.failFirstN {
		return errors.New("transient dlq publish failure")
	}
	if s.err != nil {
		return s.err
	}
	return nil
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
	h := NewHandler(reader, wallets, &stubDeadLetterPublisher{}, testLogger())

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
	h := &Handler{reader: reader, wallets: wallets, deadLetter: &stubDeadLetterPublisher{}, logger: testLogger(), retryBackoff: 0}

	msg := kafka.Message{Offset: 1, Value: mustMarshalUserCreated(t, "user-1")}
	h.processWithRetry(context.Background(), msg)

	if wallets.callCount != 3 {
		t.Fatalf("CreateWallet called %d times, want 3 (2 failures + 1 success)", wallets.callCount)
	}
	if len(reader.committed) != 1 {
		t.Fatalf("committed %d messages, want 1 — a message that eventually succeeds must still be committed", len(reader.committed))
	}
}

// TestHandler_ProcessWithRetry_GivesUpAndDeadLetters proves a message
// that never succeeds is (a) published to its dead-letter topic with the
// original value untouched plus diagnostic headers, and (b) still
// committed on the original topic afterward — a poison message must not
// block every message behind it in the same partition forever.
func TestHandler_ProcessWithRetry_GivesUpAndDeadLetters(t *testing.T) {
	reader := &stubReader{}
	wallets := &stubWalletService{err: errors.New("permanent failure")}
	dlq := &stubDeadLetterPublisher{}
	// retryBackoff: 0 — a real Handler still runs the real bounded-retry
	// loop maxProcessAttempts times, just without actually sleeping
	// between attempts (docs/CLAUDE.md §5.5: tests never rely on sleeps).
	h := &Handler{reader: reader, wallets: wallets, deadLetter: dlq, logger: testLogger(), retryBackoff: 0}

	originalValue := mustMarshalUserCreated(t, "user-1")
	msg := kafka.Message{Topic: "user.created", Partition: 0, Offset: 1, Key: []byte("user-1"), Value: originalValue}
	h.processWithRetry(context.Background(), msg)

	if wallets.callCount != maxProcessAttempts {
		t.Fatalf("CreateWallet called %d times, want exactly %d (bounded, not infinite)", wallets.callCount, maxProcessAttempts)
	}
	if len(reader.committed) != 1 {
		t.Fatalf("committed %d messages, want 1 — a permanently-failing message must still be committed after giving up, or it blocks every message behind it", len(reader.committed))
	}

	if len(dlq.publications) != 1 {
		t.Fatalf("dead-letter publisher called %d times, want 1", len(dlq.publications))
	}
	pub := dlq.publications[0]
	if pub.topic != "user.created.dlq" {
		t.Fatalf("dead-lettered to topic %q, want %q", pub.topic, "user.created.dlq")
	}
	if string(pub.value) != string(originalValue) {
		t.Fatalf("dead-lettered value = %q, want the original payload untouched: %q", pub.value, originalValue)
	}
	if string(pub.key) != "user-1" {
		t.Fatalf("dead-lettered key = %q, want the original key %q", pub.key, "user-1")
	}
	if pub.headers[headerError] == "" {
		t.Fatalf("dead-letter headers missing %q", headerError)
	}
	if pub.headers[headerAttempts] != "5" {
		t.Fatalf("dead-letter header %q = %q, want %q", headerAttempts, pub.headers[headerAttempts], "5")
	}
	if pub.headers[headerOriginalTopic] != "user.created" {
		t.Fatalf("dead-letter header %q = %q, want %q", headerOriginalTopic, pub.headers[headerOriginalTopic], "user.created")
	}
	if pub.headers[headerOriginalOffset] != "1" {
		t.Fatalf("dead-letter header %q = %q, want %q", headerOriginalOffset, pub.headers[headerOriginalOffset], "1")
	}
}

// TestHandler_ProcessWithRetry_DeadLetterPublishRecoversAfterTransientFailure
// proves the exact real-world case found live: a dead-letter topic's
// first-ever publish can fail transiently (e.g. "Unknown Topic Or
// Partition" while the broker is still creating it) and a retry
// succeeds — the message must land in the DLQ, not just get logged and
// dropped after the very first attempt.
func TestHandler_ProcessWithRetry_DeadLetterPublishRecoversAfterTransientFailure(t *testing.T) {
	reader := &stubReader{}
	wallets := &stubWalletService{err: errors.New("permanent failure")}
	dlq := &stubDeadLetterPublisher{failFirstN: 1} // fails once, succeeds on the 2nd attempt
	h := &Handler{reader: reader, wallets: wallets, deadLetter: dlq, logger: testLogger(), retryBackoff: 0}

	msg := kafka.Message{Topic: "user.created", Offset: 1, Value: mustMarshalUserCreated(t, "user-1")}
	h.processWithRetry(context.Background(), msg)

	if len(dlq.publications) != 2 {
		t.Fatalf("dead-letter publisher called %d times, want 2 (1 failure + 1 success)", len(dlq.publications))
	}
	if len(reader.committed) != 1 {
		t.Fatalf("committed %d messages, want 1", len(reader.committed))
	}
}

// TestHandler_ProcessWithRetry_SuccessNeverDeadLetters proves the happy
// path never touches the dead-letter publisher at all.
func TestHandler_ProcessWithRetry_SuccessNeverDeadLetters(t *testing.T) {
	reader := &stubReader{}
	wallets := &stubWalletService{}
	dlq := &stubDeadLetterPublisher{}
	h := &Handler{reader: reader, wallets: wallets, deadLetter: dlq, logger: testLogger(), retryBackoff: 0}

	msg := kafka.Message{Topic: "user.created", Value: mustMarshalUserCreated(t, "user-1")}
	h.processWithRetry(context.Background(), msg)

	if len(dlq.publications) != 0 {
		t.Fatalf("dead-letter publisher called %d times for a successful message, want 0", len(dlq.publications))
	}
}

// TestHandler_ProcessWithRetry_DeadLetterPublishFailureStillCommits
// proves that even if the dead-letter publish itself fails, the original
// message is still committed — the alternative (blocking the partition
// on a publish that may never succeed) is worse than the rare case this
// accepts: a message lost from both the original topic and the DLQ.
func TestHandler_ProcessWithRetry_DeadLetterPublishFailureStillCommits(t *testing.T) {
	reader := &stubReader{}
	wallets := &stubWalletService{err: errors.New("permanent failure")}
	dlq := &stubDeadLetterPublisher{err: errors.New("dlq broker unreachable")}
	h := &Handler{reader: reader, wallets: wallets, deadLetter: dlq, logger: testLogger(), retryBackoff: 0}

	msg := kafka.Message{Topic: "user.created", Offset: 1, Value: mustMarshalUserCreated(t, "user-1")}
	h.processWithRetry(context.Background(), msg)

	if len(dlq.publications) != maxDeadLetterPublishAttempts {
		t.Fatalf("dead-letter publish was attempted %d times, want exactly %d (bounded, not infinite)", len(dlq.publications), maxDeadLetterPublishAttempts)
	}
	if len(reader.committed) != 1 {
		t.Fatalf("committed %d messages, want 1 — the original message must still be committed even when the dead-letter publish fails", len(reader.committed))
	}
}

func TestHandler_Process_RejectsMalformedPayload(t *testing.T) {
	reader := &stubReader{}
	wallets := &stubWalletService{}
	h := &Handler{reader: reader, wallets: wallets, deadLetter: &stubDeadLetterPublisher{}, logger: testLogger(), retryBackoff: 0}

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
	h := &Handler{reader: reader, wallets: wallets, deadLetter: &stubDeadLetterPublisher{}, logger: testLogger(), retryBackoff: 0}

	err := h.process(context.Background(), kafka.Message{Value: mustMarshalUserCreated(t, "")})
	if err == nil {
		t.Fatalf("process() error = nil, want an error for a missing user_id")
	}
	if wallets.callCount != 0 {
		t.Fatalf("CreateWallet was called %d times for an event with no user_id, want 0", wallets.callCount)
	}
}

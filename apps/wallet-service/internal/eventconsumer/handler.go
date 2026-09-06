// Package eventconsumer processes Kafka events: fetch, handle, commit
// only after a successful (or given-up-on) handle — never before.
package eventconsumer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/encoding/protojson"

	userv1 "github.com/jrmygp/kimo-wallet/apps/wallet-service/gen/user/v1"
	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/domain"
)

// Reader is the subset of a Kafka consumer this handler needs — satisfied
// by kafkaconsumer.Consumer.
type Reader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
}

// WalletService is the subset of service.WalletService this handler needs.
type WalletService interface {
	CreateWallet(ctx context.Context, userID string) (domain.Wallet, error)
}

// DeadLetterPublisher is the subset of a Kafka producer this handler
// needs — satisfied by kafkaproducer.Producer. A message that never
// succeeds is published here, value untouched, before its original
// offset is committed — see processWithRetry.
type DeadLetterPublisher interface {
	Publish(ctx context.Context, topic string, key, value []byte, headers map[string]string) error
}

// deadLetterTopicSuffix turns the original topic into its dead-letter
// counterpart — "user.created" -> "user.created.dlq". One suffix,
// applied to whatever topic this Handler happens to be consuming, so it
// isn't hardcoded to "user.created" specifically.
const deadLetterTopicSuffix = ".dlq"

// Header keys on a dead-lettered message. The message's own Value is
// left byte-for-byte identical to the original — these headers carry
// only the diagnostic context, so a human (or a future replay tool) can
// see why it failed without needing to unwrap anything to get the
// original payload back.
const (
	headerOriginalTopic     = "x-original-topic"
	headerOriginalPartition = "x-original-partition"
	headerOriginalOffset    = "x-original-offset"
	headerFailedAt          = "x-failed-at"
	headerAttempts          = "x-attempts"
	headerError             = "x-error"
)

// maxProcessAttempts bounds retry of one message before giving up on it
// and dead-lettering it — never an infinite *tight* loop (docs/CLAUDE.md
// §3.6 rule 3).
const maxProcessAttempts = 5

// maxDeadLetterPublishAttempts bounds retry of the dead-letter publish
// itself, separately from maxProcessAttempts. Found live: a dead-letter
// topic's very first publish reliably fails with "Unknown Topic Or
// Partition" even with the producer's AllowAutoTopicCreation set — the
// broker starts creating the topic as a side effect of that first
// request, but creation isn't synchronous with it, so that same request
// still reports the topic missing. A short retry clears this on the
// first dead-letter ever sent to a given topic; every one after that
// succeeds on the first attempt, since the topic already exists by then.
const maxDeadLetterPublishAttempts = 3

// defaultRetryBackoff is NewHandler's production value. A field, not a
// baked-in constant used directly in Run/processWithRetry, specifically
// so tests can construct a Handler with zero backoff — this package's
// tests must exercise the real bounded-retry-then-commit loop without
// actually sleeping (docs/CLAUDE.md §5.5: tests never rely on sleeps).
const defaultRetryBackoff = 2 * time.Second

type Handler struct {
	reader       Reader
	wallets      WalletService
	deadLetter   DeadLetterPublisher
	logger       *slog.Logger
	retryBackoff time.Duration
}

func NewHandler(reader Reader, wallets WalletService, deadLetter DeadLetterPublisher, logger *slog.Logger) *Handler {
	return &Handler{reader: reader, wallets: wallets, deadLetter: deadLetter, logger: logger, retryBackoff: defaultRetryBackoff}
}

// Run fetches and processes messages one at a time until ctx is
// cancelled. Single-threaded — sufficient for this volume; revisit only
// if there's an actual throughput reason to process concurrently.
func (h *Handler) Run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		msg, err := h.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			h.logger.Error("fetch kafka message", "error", err.Error())
			time.Sleep(h.retryBackoff)
			continue
		}

		h.processWithRetry(ctx, msg)
	}
}

func (h *Handler) processWithRetry(ctx context.Context, msg kafka.Message) {
	var lastErr error
	for attempt := 1; attempt <= maxProcessAttempts; attempt++ {
		lastErr = h.process(ctx, msg)
		if lastErr == nil {
			break
		}
		h.logger.Error("process user.created event", "error", lastErr.Error(), "attempt", attempt, "offset", msg.Offset)
		if attempt < maxProcessAttempts {
			time.Sleep(h.retryBackoff)
		}
	}

	if lastErr != nil {
		h.logger.Error("giving up on event after max attempts, dead-lettering",
			"attempts", maxProcessAttempts, "offset", msg.Offset, "error", lastErr.Error())
		h.deadLetterOrLog(ctx, msg, lastErr)
	}

	// Committed either way: a message that fails forever (dead-lettered
	// or not) must not block every message behind it in the same partition.
	if err := h.reader.CommitMessages(ctx, msg); err != nil {
		h.logger.Error("commit kafka message", "error", err.Error(), "offset", msg.Offset)
	}
}

// deadLetterOrLog publishes msg, value untouched, to its dead-letter
// topic — see the header constants above for what travels alongside it.
// If the dead-letter publish itself fails (e.g. Kafka briefly
// unreachable at that exact moment), that's logged loudly and the
// original message is still committed regardless: the alternative is
// blocking the whole partition indefinitely on a publish that may never
// succeed, which is worse than the rare double failure this accepts.
func (h *Handler) deadLetterOrLog(ctx context.Context, msg kafka.Message, processErr error) {
	headers := map[string]string{
		headerOriginalTopic:     msg.Topic,
		headerOriginalPartition: strconv.Itoa(msg.Partition),
		headerOriginalOffset:    strconv.FormatInt(msg.Offset, 10),
		headerFailedAt:          time.Now().UTC().Format(time.RFC3339),
		headerAttempts:          strconv.Itoa(maxProcessAttempts),
		headerError:             processErr.Error(),
	}

	dlqTopic := msg.Topic + deadLetterTopicSuffix

	var publishErr error
	for attempt := 1; attempt <= maxDeadLetterPublishAttempts; attempt++ {
		publishErr = h.deadLetter.Publish(ctx, dlqTopic, msg.Key, msg.Value, headers)
		if publishErr == nil {
			return
		}
		if attempt < maxDeadLetterPublishAttempts {
			time.Sleep(h.retryBackoff)
		}
	}

	h.logger.Error("publish to dead-letter topic failed after retries — message content will only be recoverable from the original topic's own retention, if any",
		"dlq_topic", dlqTopic, "offset", msg.Offset, "error", publishErr.Error())
}

func (h *Handler) process(ctx context.Context, msg kafka.Message) error {
	var event userv1.UserCreated
	if err := protojson.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal user.created payload: %w", err)
	}
	if event.GetUserId() == "" {
		return errors.New("user.created event missing user_id")
	}

	if _, err := h.wallets.CreateWallet(ctx, event.GetUserId()); err != nil {
		return fmt.Errorf("create wallet for user %s: %w", event.GetUserId(), err)
	}
	return nil
}

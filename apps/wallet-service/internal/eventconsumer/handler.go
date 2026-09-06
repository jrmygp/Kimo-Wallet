// Package eventconsumer processes Kafka events: fetch, handle, commit
// only after a successful (or given-up-on) handle — never before.
package eventconsumer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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

// maxProcessAttempts bounds retry of one message before giving up on it
// — never an infinite *tight* loop (docs/CLAUDE.md §3.6 rule 3). There's
// no dead-letter *topic* yet — "giving up" means a loud log and moving
// on to the next message, not blocking the whole partition behind one
// bad event forever. That's a real gap, not a silent one — see this
// day's log entry.
const maxProcessAttempts = 5

// defaultRetryBackoff is NewHandler's production value. A field, not a
// baked-in constant used directly in Run/processWithRetry, specifically
// so tests can construct a Handler with zero backoff — this package's
// tests must exercise the real bounded-retry-then-commit loop without
// actually sleeping (docs/CLAUDE.md §5.5: tests never rely on sleeps).
const defaultRetryBackoff = 2 * time.Second

type Handler struct {
	reader       Reader
	wallets      WalletService
	logger       *slog.Logger
	retryBackoff time.Duration
}

func NewHandler(reader Reader, wallets WalletService, logger *slog.Logger) *Handler {
	return &Handler{reader: reader, wallets: wallets, logger: logger, retryBackoff: defaultRetryBackoff}
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
		h.logger.Error("giving up on event after max attempts — skipping, not dead-lettering (known gap)",
			"attempts", maxProcessAttempts, "offset", msg.Offset, "error", lastErr.Error())
	}

	// Committed either way: a message that fails forever must not block
	// every message behind it in the same partition.
	if err := h.reader.CommitMessages(ctx, msg); err != nil {
		h.logger.Error("commit kafka message", "error", err.Error(), "offset", msg.Offset)
	}
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

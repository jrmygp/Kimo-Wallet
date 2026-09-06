// Package outbox relays transactional outbox rows to Kafka: repeatedly
// polling for unpublished events and publishing them, marking each one
// published only after a successful publish.
package outbox

import (
	"context"
	"log/slog"
	"time"

	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/domain"
)

// Repository is the subset of storage the relay needs — satisfied by
// postgres.OutboxRepository.
type Repository interface {
	FetchUnpublished(ctx context.Context, limit int) ([]domain.OutboxEvent, error)
	MarkPublished(ctx context.Context, id string) error
}

// Publisher is the subset of a Kafka client the relay needs — satisfied
// by kafkaproducer.Producer.
type Publisher interface {
	Publish(ctx context.Context, topic string, key, value []byte) error
}

// batchSize bounds how many unpublished events one tick fetches, so a
// large backlog (e.g. after Kafka being down for a while) doesn't load an
// unbounded result set in one query.
const batchSize = 100

type Relay struct {
	repo      Repository
	publisher Publisher
	interval  time.Duration
	logger    *slog.Logger
}

func NewRelay(repo Repository, publisher Publisher, interval time.Duration, logger *slog.Logger) *Relay {
	return &Relay{repo: repo, publisher: publisher, interval: interval, logger: logger}
}

// Run polls on a fixed interval until ctx is cancelled. A failed publish
// is retried on the next tick, not immediately — bounded by the ticker,
// never a tight retry loop (docs/CLAUDE.md §3.6 rule 3). There's no
// "poison pill" concern to dead-letter here the way an event *consumer*
// would have: an unpublished row has nothing wrong with it, it just needs
// Kafka to be reachable, so retrying indefinitely on a timer is correct,
// not a bug.
func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.relayOnce(ctx)
		}
	}
}

func (r *Relay) relayOnce(ctx context.Context) {
	events, err := r.repo.FetchUnpublished(ctx, batchSize)
	if err != nil {
		r.logger.Error("fetch unpublished outbox events", "error", err.Error())
		return
	}

	for _, event := range events {
		// event_type doubles as the Kafka topic name — a deliberate 1:1
		// mapping, see docs/guides/Kimo-Wallet-Architecture.md §5's event list.
		if err := r.publisher.Publish(ctx, event.EventType, []byte(event.ID), event.Payload); err != nil {
			r.logger.Error("publish outbox event", "error", err.Error(), "event_id", event.ID, "event_type", event.EventType)
			continue
		}

		if err := r.repo.MarkPublished(ctx, event.ID); err != nil {
			// Published but not marked: next tick republishes the same
			// event. Accepted at-least-once duplicate — consumers must
			// already dedupe (docs/CLAUDE.md §3.6 rule 1).
			r.logger.Error("mark outbox event published", "error", err.Error(), "event_id", event.ID)
		}
	}
}

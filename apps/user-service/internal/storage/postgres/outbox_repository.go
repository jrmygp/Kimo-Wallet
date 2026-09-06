package postgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/domain"
)

// EventTypeUserCreated is the outbox event_type / Kafka topic name for a
// newly registered user. Topic name and event_type are the same string —
// see outbox.Relay.
const EventTypeUserCreated = "user.created"

// outboxEventModel is the GORM row mapping for the outbox_events table.
// Payload is stored (and republished) as opaque bytes — this layer never
// inspects its contents; see user_repository.go's Create for what builds it.
type outboxEventModel struct {
	ID          string `gorm:"primaryKey"`
	EventType   string
	Payload     []byte
	CreatedAt   time.Time
	PublishedAt *time.Time
}

func (outboxEventModel) TableName() string { return "outbox_events" }

type OutboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

// FetchUnpublished returns up to limit not-yet-published events, oldest
// first, so the relay publishes them in the order they were created.
func (r *OutboxRepository) FetchUnpublished(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	var rows []outboxEventModel

	if err := r.db.WithContext(ctx).
		Where("published_at IS NULL").
		Order("created_at ASC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("fetch unpublished outbox events: %w", err)
	}

	events := make([]domain.OutboxEvent, len(rows))
	for i, row := range rows {
		events[i] = domain.OutboxEvent{ID: row.ID, EventType: row.EventType, Payload: row.Payload}
	}
	return events, nil
}

// MarkPublished records that id has been handed to Kafka. Called only
// after a successful publish; if this fails after that publish succeeded,
// the event is republished on the next tick — an accepted at-least-once
// duplicate, per docs/CLAUDE.md §3.6 rule 1 (consumers must dedupe anyway).
func (r *OutboxRepository) MarkPublished(ctx context.Context, id string) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&outboxEventModel{}).
		Where("id = ?", id).
		Update("published_at", now).Error; err != nil {
		return fmt.Errorf("mark outbox event %s published: %w", id, err)
	}
	return nil
}

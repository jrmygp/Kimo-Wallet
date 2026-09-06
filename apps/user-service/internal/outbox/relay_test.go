package outbox

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/domain"
)

type stubRepository struct {
	unpublished    []domain.OutboxEvent
	fetchErr       error
	markPublishErr error
	published      []string // ids MarkPublished was called with
}

func (s *stubRepository) FetchUnpublished(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	if s.fetchErr != nil {
		return nil, s.fetchErr
	}
	return s.unpublished, nil
}

func (s *stubRepository) MarkPublished(ctx context.Context, id string) error {
	if s.markPublishErr != nil {
		return s.markPublishErr
	}
	s.published = append(s.published, id)
	return nil
}

type stubPublisher struct {
	publishErr error
	published  []string // ids Publish was called with (via the message key)
}

func (s *stubPublisher) Publish(ctx context.Context, topic string, key, value []byte) error {
	if s.publishErr != nil {
		return s.publishErr
	}
	s.published = append(s.published, string(key))
	return nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRelay_RelayOnce_PublishesAndMarksEachUnpublishedEvent(t *testing.T) {
	repo := &stubRepository{unpublished: []domain.OutboxEvent{
		{ID: "event-1", EventType: "user.created", Payload: []byte(`{"userId":"1"}`)},
		{ID: "event-2", EventType: "user.created", Payload: []byte(`{"userId":"2"}`)},
	}}
	publisher := &stubPublisher{}

	relay := NewRelay(repo, publisher, 0, testLogger())
	relay.relayOnce(context.Background())

	if len(publisher.published) != 2 || publisher.published[0] != "event-1" || publisher.published[1] != "event-2" {
		t.Fatalf("publisher got %v, want [event-1 event-2]", publisher.published)
	}
	if len(repo.published) != 2 || repo.published[0] != "event-1" || repo.published[1] != "event-2" {
		t.Fatalf("repo.MarkPublished got %v, want [event-1 event-2]", repo.published)
	}
}

func TestRelay_RelayOnce_PublishFailureLeavesEventUnmarked(t *testing.T) {
	repo := &stubRepository{unpublished: []domain.OutboxEvent{
		{ID: "event-1", EventType: "user.created", Payload: []byte(`{}`)},
	}}
	publisher := &stubPublisher{publishErr: errors.New("broker unreachable")}

	relay := NewRelay(repo, publisher, 0, testLogger())
	relay.relayOnce(context.Background())

	if len(repo.published) != 0 {
		t.Fatalf("MarkPublished was called %v after a failed Publish, want it never called — the event must stay unpublished for the next tick to retry", repo.published)
	}
}

func TestRelay_RelayOnce_FetchFailureDoesNotPanic(t *testing.T) {
	repo := &stubRepository{fetchErr: errors.New("db unreachable")}
	publisher := &stubPublisher{}

	relay := NewRelay(repo, publisher, 0, testLogger())
	relay.relayOnce(context.Background()) // must not panic

	if len(publisher.published) != 0 {
		t.Fatalf("publisher was called %v despite FetchUnpublished failing", publisher.published)
	}
}

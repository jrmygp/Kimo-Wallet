-- Transactional outbox: an event row is written in the same DB transaction
-- as the effect that produced it (see internal/storage/postgres/user_repository.go's
-- Create), so the event can never be lost even if the process dies right
-- after the commit. A separate relay (internal/outbox) polls unpublished
-- rows and publishes them to Kafka, then marks them published.
CREATE TABLE outbox_events (
    id           UUID PRIMARY KEY,
    event_type   TEXT NOT NULL,
    payload      JSONB NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ
);

-- The relay's only query is "give me unpublished rows, oldest first" —
-- a partial index on that predicate keeps it cheap regardless of how many
-- already-published rows have piled up.
CREATE INDEX outbox_events_unpublished_idx ON outbox_events (created_at) WHERE published_at IS NULL;

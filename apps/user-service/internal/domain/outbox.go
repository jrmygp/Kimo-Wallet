package domain

// OutboxEvent is one row of the transactional outbox: an event recorded
// in the same DB transaction as the effect that produced it, later
// relayed to Kafka. See internal/storage/postgres/outbox_repository.go
// (writer, via user_repository.go's Create) and internal/outbox (reader
// and publisher).
type OutboxEvent struct {
	ID        string
	EventType string
	Payload   []byte
}

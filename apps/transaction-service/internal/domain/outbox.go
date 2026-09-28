package domain

type OutboxEvent struct {
	ID        string
	EventType string
	Payload   []byte
}

// Package kafkaproducer wraps segmentio/kafka-go behind the small
// interface internal/outbox actually needs, so nothing outside this
// package touches the client library directly.
package kafkaproducer

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

// Producer publishes messages to Kafka topics.
type Producer struct {
	writer *kafka.Writer
}

func New(brokers []string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Balancer: &kafka.LeastBytes{},
			// Without this, kafka-go treats "topic doesn't exist yet" as a
			// hard error rather than asking the broker to create it. This
			// service's outbox relay retries every 2s regardless, so the
			// symptom self-healed silently here — but it's the same latent
			// gap found live in wallet-service's dead-letter producer,
			// which only publishes once per message and doesn't get that
			// same free retry.
			AllowAutoTopicCreation: true,
		},
	}
}

// Publish sends value to topic, keyed by key. Topic is per-message, not
// fixed on the writer, since this one Producer serves every outbox event
// type — see internal/outbox.Relay.
func (p *Producer) Publish(ctx context.Context, topic string, key, value []byte) error {
	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	}); err != nil {
		return fmt.Errorf("publish to kafka topic %s: %w", topic, err)
	}
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

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

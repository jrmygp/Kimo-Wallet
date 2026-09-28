// Package kafkaproducer wraps segmentio/kafka-go behind the small
// interface internal/eventconsumer actually needs (publishing to the
// dead-letter topic), so nothing outside this package touches the
// client library directly. Duplicated from user-service's own
// kafkaproducer package rather than shared — separate Go modules — with
// headers support added, since the dead-letter use case needs it and
// the outbox relay's doesn't.
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
			// hard error rather than asking the broker to create it —
			// found live: a dead-letter topic's very first publish failed
			// with "Unknown Topic Or Partition" because of this default.
			AllowAutoTopicCreation: true,
		},
	}
}

// Publish sends value to topic, keyed by key, with headers attached as
// plain string key/value pairs. Topic is per-message, not fixed on the
// writer, since this one Producer could serve more than one destination.
func (p *Producer) Publish(ctx context.Context, topic string, key, value []byte, headers map[string]string) error {
	kafkaHeaders := make([]kafka.Header, 0, len(headers))
	for k, v := range headers {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: k, Value: []byte(v)})
	}

	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Topic:   topic,
		Key:     key,
		Value:   value,
		Headers: kafkaHeaders,
	}); err != nil {
		return fmt.Errorf("publish to kafka topic %s: %w", topic, err)
	}
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

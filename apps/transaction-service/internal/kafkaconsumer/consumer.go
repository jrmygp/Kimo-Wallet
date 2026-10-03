// Package kafkaconsumer wraps segmentio/kafka-go's consumer-group Reader.
package kafkaconsumer

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// Consumer reads from one topic as part of a named consumer group.
// Offsets are never auto-committed — FetchMessage (not ReadMessage, which
// would auto-commit) leaves committing to the caller, via CommitMessages,
// so a message is only ever marked done after it's actually been
// processed. See internal/eventconsumer.Handler, which owns that decision.
type Consumer struct {
	reader *kafka.Reader
}

func New(brokers []string, groupID, topic string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			GroupID: groupID,
			Topic:   topic,
		}),
	}
}

func (c *Consumer) FetchMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.FetchMessage(ctx)
}

func (c *Consumer) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	return c.reader.CommitMessages(ctx, msgs...)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

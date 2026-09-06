// Package config loads wallet-service configuration from the environment.
package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL        string
	KafkaBrokers       []string
	KafkaConsumerGroup string
}

// Load reads configuration from environment variables, failing fast if a
// required value is missing rather than falling back to a guessed default
// for anything that affects where data is stored.
func Load() (Config, error) {
	databaseURL := os.Getenv("WALLET_SERVICE_DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("WALLET_SERVICE_DATABASE_URL is required")
	}

	// Soft-defaulted, same reasoning as user-service's KafkaBrokers: the
	// local dev broker (docker-compose.yml's kafka service), overridable
	// per environment.
	kafkaBrokers := os.Getenv("WALLET_SERVICE_KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	// Consumer group id — distinct from any other service's, so Kafka
	// tracks this service's own committed offset per topic independently.
	kafkaConsumerGroup := os.Getenv("WALLET_SERVICE_KAFKA_CONSUMER_GROUP")
	if kafkaConsumerGroup == "" {
		kafkaConsumerGroup = "wallet-service"
	}

	return Config{
		DatabaseURL:        databaseURL,
		KafkaBrokers:       strings.Split(kafkaBrokers, ","),
		KafkaConsumerGroup: kafkaConsumerGroup,
	}, nil
}

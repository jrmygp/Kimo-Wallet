// Package config loads user-service configuration from the environment.
package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	GRPCPort     string
	DatabaseURL  string
	JWTSecret    []byte
	KafkaBrokers []string
}

// Load reads configuration from environment variables, failing fast if a
// required value is missing rather than falling back to a guessed default
// for anything that affects where data is stored or how tokens are signed.
func Load() (Config, error) {
	databaseURL := os.Getenv("USER_SERVICE_DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("USER_SERVICE_DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	grpcPort := os.Getenv("USER_SERVICE_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	// Defaults to the local dev broker (docker-compose.yml's kafka
	// service). Outbox publishing failing/retrying is expected to be
	// tolerated, not fatal (see internal/outbox), so this — unlike
	// DatabaseURL/JWTSecret — has a fallback rather than failing fast.
	kafkaBrokers := os.Getenv("USER_SERVICE_KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	return Config{
		GRPCPort:     grpcPort,
		DatabaseURL:  databaseURL,
		JWTSecret:    []byte(jwtSecret),
		KafkaBrokers: strings.Split(kafkaBrokers, ","),
	}, nil
}

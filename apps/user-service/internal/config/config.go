// Package config loads user-service configuration from the environment.
package config

import (
	"fmt"
	"os"
)

type Config struct {
	GRPCPort    string
	DatabaseURL string
}

// Load reads configuration from environment variables, failing fast if a
// required value is missing rather than falling back to a guessed default
// for anything that affects where data is stored.
func Load() (Config, error) {
	databaseURL := os.Getenv("USER_SERVICE_DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("USER_SERVICE_DATABASE_URL is required")
	}

	grpcPort := os.Getenv("USER_SERVICE_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	return Config{
		GRPCPort:    grpcPort,
		DatabaseURL: databaseURL,
	}, nil
}

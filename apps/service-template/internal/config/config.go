// Package config loads service-template configuration from the
// environment.
//
// TEMPLATE NOTE: rename every SERVICE_TEMPLATE_* env var below to your
// service's own prefix (e.g. WIDGET_SERVICE_*) — see README.md.
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
	databaseURL := os.Getenv("SERVICE_TEMPLATE_DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("SERVICE_TEMPLATE_DATABASE_URL is required")
	}

	grpcPort := os.Getenv("SERVICE_TEMPLATE_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50053"
	}

	return Config{
		GRPCPort:    grpcPort,
		DatabaseURL: databaseURL,
	}, nil
}

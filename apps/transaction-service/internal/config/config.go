// Package config loads transaction-service configuration from the
// environment.
package config

import (
	"fmt"
	"os"
)

type Config struct {
	GRPCPort          string
	DatabaseURL       string
	WalletServiceAddr string
}

// Load reads configuration from environment variables, failing fast if a
// required value is missing rather than falling back to a guessed default
// for anything that affects where data is stored.
func Load() (Config, error) {
	databaseURL := os.Getenv("TRANSACTION_SERVICE_DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("TRANSACTION_SERVICE_DATABASE_URL is required")
	}

	grpcPort := os.Getenv("TRANSACTION_SERVICE_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50053"
	}

	walletServiceAddr := os.Getenv("WALLET_SERVICE_GRPC_ADDR")
	if walletServiceAddr == "" {
		return Config{}, fmt.Errorf("WALLET_SERVICE_GRPC_ADDR is required")
	}

	return Config{
		GRPCPort:          grpcPort,
		DatabaseURL:       databaseURL,
		WalletServiceAddr: walletServiceAddr,
	}, nil
}

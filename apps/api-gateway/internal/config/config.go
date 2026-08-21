// Package config loads api-gateway configuration from the environment.
package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPPort        string
	UserServiceAddr string
}

// Load reads configuration from environment variables, failing fast if a
// required value is missing rather than falling back to a guessed default
// for where an upstream service lives.
func Load() (Config, error) {
	userServiceAddr := os.Getenv("USER_SERVICE_GRPC_ADDR")
	if userServiceAddr == "" {
		return Config{}, fmt.Errorf("USER_SERVICE_GRPC_ADDR is required")
	}

	httpPort := os.Getenv("API_GATEWAY_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	return Config{
		HTTPPort:        httpPort,
		UserServiceAddr: userServiceAddr,
	}, nil
}

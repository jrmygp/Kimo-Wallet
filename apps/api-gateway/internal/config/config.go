// Package config loads api-gateway configuration from the environment.
package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPPort          string
	UserServiceAddr   string
	JWTSecret         []byte
	CORSAllowedOrigin string
}

// Load reads configuration from environment variables, failing fast if a
// required value is missing rather than falling back to a guessed default
// for where an upstream service lives or how tokens are verified.
func Load() (Config, error) {
	userServiceAddr := os.Getenv("USER_SERVICE_GRPC_ADDR")
	if userServiceAddr == "" {
		return Config{}, fmt.Errorf("USER_SERVICE_GRPC_ADDR is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	httpPort := os.Getenv("API_GATEWAY_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	corsAllowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if corsAllowedOrigin == "" {
		corsAllowedOrigin = "http://localhost:3000"
	}

	return Config{
		HTTPPort:          httpPort,
		UserServiceAddr:   userServiceAddr,
		JWTSecret:         []byte(jwtSecret),
		CORSAllowedOrigin: corsAllowedOrigin,
	}, nil
}

package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/jrmygp/kimo-wallet/apps/api-gateway/internal/config"
	"github.com/jrmygp/kimo-wallet/apps/api-gateway/internal/handler"
	"github.com/jrmygp/kimo-wallet/apps/api-gateway/internal/httpserver"

	userv1 "github.com/jrmygp/kimo-wallet/apps/api-gateway/gen/user/v1"
)

const serviceName = "api-gateway"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", serviceName)

	if err := run(logger); err != nil {
		logger.Error("service exited with error", "error", err.Error())
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// .env is a local-dev convenience (gitignored, loaded from the working
	// directory) and never overrides a variable already set in the real
	// environment. Its absence is normal in production/CI/Docker.
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Insecure transport credentials: this connects to user-service over a
	// local/private network (localhost in dev, the compose/cluster network
	// in deployment) with no TLS yet. Per docs/CLAUDE.md §19 the deployed
	// architecture requires TLS — tracked as a known gap, not silently
	// skipped, since this is still the local/dev stage of the project.
	userServiceConn, err := grpc.NewClient(cfg.UserServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("connect to user-service: %w", err)
	}
	defer userServiceConn.Close()

	userClient := userv1.NewUserServiceClient(userServiceConn)
	userHandler := handler.NewUserHandler(userClient)
	router := httpserver.NewRouter(userHandler)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "port", cfg.HTTPPort, "user_service_addr", cfg.UserServiceAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received, stopping gracefully")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}

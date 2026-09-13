package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	widgetv1 "github.com/jrmygp/kimo-wallet/apps/service-template/gen/widget/v1"
	"github.com/jrmygp/kimo-wallet/apps/service-template/internal/config"
	"github.com/jrmygp/kimo-wallet/apps/service-template/internal/grpcserver"
	"github.com/jrmygp/kimo-wallet/apps/service-template/internal/service"
	"github.com/jrmygp/kimo-wallet/apps/service-template/internal/storage/postgres"
	"github.com/jrmygp/kimo-wallet/apps/service-template/migrations"
)

// TEMPLATE NOTE: rename this and everywhere it's used to your service's
// real name — see README.md.
const serviceName = "service-template"

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
	// environment — see joho/godotenv's Load() docs. Its absence is normal
	// in production/CI/Docker; only a malformed .env that IS present is a
	// real error. Same convention as every other service in this repo.
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := postgres.Open(cfg.DatabaseURL)
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB: %w", err)
	}
	defer sqlDB.Close()

	if err := postgres.Migrate(ctx, sqlDB, migrations.Files); err != nil {
		return err
	}
	logger.Info("migrations applied")

	repo := postgres.NewWidgetRepository(db)
	widgetService := service.NewWidgetService(repo)
	widgetServer := grpcserver.NewWidgetServer(widgetService)

	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	widgetv1.RegisterWidgetServiceServer(grpcServer, widgetServer)

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("grpc server listening", "port", cfg.GRPCPort)
		serveErr <- grpcServer.Serve(listener)
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received, stopping gracefully")
		grpcServer.GracefulStop()
		return nil
	}
}

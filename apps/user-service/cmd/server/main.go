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
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/authtoken"
	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/config"
	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/grpcserver"
	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/kafkaproducer"
	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/outbox"
	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/service"
	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/storage/postgres"
	"github.com/jrmygp/kimo-wallet/apps/user-service/migrations"

	userv1 "github.com/jrmygp/kimo-wallet/apps/user-service/gen/user/v1"
)

const serviceName = "user-service"

// outboxRelayInterval is how often the outbox relay polls for unpublished
// events. Not tunable via env yet — a fixed, boring interval is fine for
// this phase; see internal/outbox for why a bounded-not-tight retry on a
// timer is the correct behavior here, not a workaround.
const outboxRelayInterval = 2 * time.Second

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
	// in production/CI/Docker, where config comes from the real environment
	// instead; only a malformed .env that IS present is a real error.
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

	repo := postgres.NewUserRepository(db)
	minter := authtoken.NewMinter(cfg.JWTSecret)
	userService := service.NewUserService(repo, minter)
	userServer := grpcserver.NewUserServer(userService)

	producer := kafkaproducer.New(cfg.KafkaBrokers)
	defer func() {
		if err := producer.Close(); err != nil {
			logger.Error("close kafka producer", "error", err.Error())
		}
	}()

	outboxRepo := postgres.NewOutboxRepository(db)
	relay := outbox.NewRelay(outboxRepo, producer, outboxRelayInterval, logger)
	go relay.Run(ctx)
	logger.Info("outbox relay started", "interval", outboxRelayInterval.String())

	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	userv1.RegisterUserServiceServer(grpcServer, userServer)

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

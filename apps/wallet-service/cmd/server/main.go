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

	walletv1 "github.com/jrmygp/kimo-wallet/apps/wallet-service/gen/wallet/v1"
	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/config"
	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/eventconsumer"
	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/grpcserver"
	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/kafkaconsumer"
	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/kafkaproducer"
	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/service"
	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/storage/postgres"
	"github.com/jrmygp/kimo-wallet/apps/wallet-service/migrations"
)

const serviceName = "wallet-service"

// userCreatedTopic must match user-service's EventTypeUserCreated
// (apps/user-service/internal/storage/postgres/outbox_repository.go) —
// the two services don't share Go code, so this string is the actual
// contract between them, not the .proto (which only defines the
// message shape, not the topic name).
const userCreatedTopic = "user.created"

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
	// real error. Same convention as user-service's main.go.
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

	repo := postgres.NewWalletRepository(db)
	walletService := service.NewWalletService(repo)
	walletServer := grpcserver.NewWalletServer(walletService)

	consumer := kafkaconsumer.New(cfg.KafkaBrokers, cfg.KafkaConsumerGroup, userCreatedTopic)
	defer func() {
		if err := consumer.Close(); err != nil {
			logger.Error("close kafka consumer", "error", err.Error())
		}
	}()

	// Same broker list as the consumer — this producer only ever writes
	// to <topic>.dlq when Handler gives up on a message, see
	// internal/eventconsumer's Handler.deadLetterOrLog.
	deadLetterProducer := kafkaproducer.New(cfg.KafkaBrokers)
	defer func() {
		if err := deadLetterProducer.Close(); err != nil {
			logger.Error("close dead-letter kafka producer", "error", err.Error())
		}
	}()

	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	walletv1.RegisterWalletServiceServer(grpcServer, walletServer)

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("grpc server listening", "port", cfg.GRPCPort)
		serveErr <- grpcServer.Serve(listener)
	}()

	handler := eventconsumer.NewHandler(consumer, walletService, deadLetterProducer, logger)
	go func() {
		logger.Info("consuming", "topic", userCreatedTopic, "group", cfg.KafkaConsumerGroup)
		handler.Run(ctx)
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		logger.Info("consuming", "topic", userCreatedTopic, "group", cfg.KafkaConsumerGroup)
		logger.Info("shutdown signal received, stopped consuming")
		return nil
	}
}

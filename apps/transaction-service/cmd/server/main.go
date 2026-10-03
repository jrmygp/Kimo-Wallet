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
	"google.golang.org/grpc/credentials/insecure"

	transactionv1 "github.com/jrmygp/kimo-wallet/apps/transaction-service/gen/transaction/v1"
	walletv1 "github.com/jrmygp/kimo-wallet/apps/transaction-service/gen/wallet/v1"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/config"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/eventconsumer"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/grpcserver"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/kafkaconsumer"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/kafkaproducer"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/outbox"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/service"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/storage/postgres"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/migrations"
)

const serviceName = "transaction-service"

const transactionProcessedTopic = "transaction.processed"

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

	// Wallet service grpc connection — used to confirm the sender/receiver
	// actually have a wallet before creating a transaction (see
	// internal/service/transaction_service.go). No separate connection to
	// user-service: a wallet can only exist for a real user (created via
	// wallet-service's user.created consumer), so a successful wallet
	// lookup already proves the user exists too.
	walletServiceConn, err := grpc.NewClient(cfg.WalletServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("connect to wallet-service: %w", err)
	}
	defer walletServiceConn.Close()

	walletClient := walletv1.NewWalletServiceClient(walletServiceConn)

	repo := postgres.NewTransactionRepository(db)
	transactionService := service.NewTransactionService(repo, walletClient)
	transactionServer := grpcserver.NewTransactionServer(transactionService)

	walletConsumer := kafkaconsumer.New(cfg.KafkaBrokers, cfg.KafkaConsumerGroup, transactionProcessedTopic)
	defer func() {
		if err := walletConsumer.Close(); err != nil {
			logger.Error("close kafka consumer", "error", err.Error())
		}
	}()

	deadLetterProducer := kafkaproducer.New(cfg.KafkaBrokers)
	defer func() {
		if err := deadLetterProducer.Close(); err != nil {
			logger.Error("close dead-letter kafka producer", "error", err.Error())
		}
	}()

	producer := kafkaproducer.New(cfg.KafkaBrokers)

	outboxRepo := postgres.NewOutboxRepository(db)
	relay := outbox.NewRelay(outboxRepo, producer, outboxRelayInterval, logger)
	go relay.Run(ctx)
	logger.Info("outbox relay started", "interval", outboxRelayInterval.String())

	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	transactionv1.RegisterTransactionServiceServer(grpcServer, transactionServer)

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("grpc server listening", "port", cfg.GRPCPort)
		serveErr <- grpcServer.Serve(listener)
	}()

	walletHandler := eventconsumer.NewHandler(walletConsumer, eventconsumer.NewTransactionProcessedProcessor(transactionService), deadLetterProducer, logger)
	go func() {
		logger.Info("consuming", "topic", transactionProcessedTopic, "group", cfg.KafkaConsumerGroup)
		walletHandler.Run(ctx)
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

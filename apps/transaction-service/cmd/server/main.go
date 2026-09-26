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
	"google.golang.org/grpc/credentials/insecure"

	transactionv1 "github.com/jrmygp/kimo-wallet/apps/transaction-service/gen/transaction/v1"
	walletv1 "github.com/jrmygp/kimo-wallet/apps/transaction-service/gen/wallet/v1"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/config"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/grpcserver"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/service"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/internal/storage/postgres"
	"github.com/jrmygp/kimo-wallet/apps/transaction-service/migrations"
)

// TEMPLATE NOTE: rename this and everywhere it's used to your service's
// real name — see README.md.
const serviceName = "transaction-service"

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

	// User service grpc connection
	// userServiceConn, err := grpc.NewClient(cfg.UserServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	return fmt.Errorf("connect to user-service: %w", err)
	// }
	// defer userServiceConn.Close()

	// Wallet service grpc connection
	walletServiceConn, err := grpc.NewClient(cfg.WalletServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("connect to wallet-service: %w", err)
	}
	defer walletServiceConn.Close()

	walletClient := walletv1.NewWalletServiceClient(walletServiceConn)

	repo := postgres.NewTransactionRepository(db)
	transactionService := service.NewTransactionService(repo, walletClient)
	transactionServer := grpcserver.NewTransactionServer(transactionService)

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

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received, stopping gracefully")
		grpcServer.GracefulStop()
		return nil
	}
}

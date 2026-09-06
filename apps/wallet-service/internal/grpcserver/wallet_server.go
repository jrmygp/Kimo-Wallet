package grpcserver

import (
	"context"
	"errors"

	walletv1 "github.com/jrmygp/kimo-wallet/apps/wallet-service/gen/wallet/v1"
	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type walletService interface {
	GetWalletByUserID(ctx context.Context, userID string) (domain.Wallet, error)
}

func toProtoWallet(wallet domain.Wallet) *walletv1.Wallet {
	return &walletv1.Wallet{
		Id:        wallet.ID,
		UserId:    wallet.UserID,
		Balance:   wallet.Balance,
		Currency:  wallet.Currency,
		Status:    wallet.Status,
		CreatedAt: timestamppb.New(wallet.CreatedAt),
	}
}

type WalletServer struct {
	walletv1.UnimplementedWalletServiceServer
	service walletService
}

func NewWalletServer(service walletService) *WalletServer {
	return &WalletServer{service: service}
}

func (s *WalletServer) GetWalletByUserID(ctx context.Context, req *walletv1.GetWalletByUserIDRequest) (*walletv1.GetWalletByUserIDResponse, error) {
	wallet, err := s.service.GetWalletByUserID(ctx, req.GetUserId())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidUserID):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, domain.ErrWalletNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to find wallet")
		}
	}

	return &walletv1.GetWalletByUserIDResponse{
		Wallet: toProtoWallet(wallet),
	}, nil
}

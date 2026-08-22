// Package grpcserver adapts the UserService domain logic to the generated
// gRPC transport: translating domain errors to gRPC status codes and back.
package grpcserver

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	userv1 "github.com/jrmygp/kimo-wallet/apps/user-service/gen/user/v1"
	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/domain"
)

// userService is the subset of service.UserService this server depends on.
type userService interface {
	Register(ctx context.Context, phoneNumber, fullName string) (domain.User, error)
	Login(ctx context.Context, phoneNumber string) (domain.User, error)
}

type UserServer struct {
	userv1.UnimplementedUserServiceServer
	service userService
}

func NewUserServer(service userService) *UserServer {
	return &UserServer{service: service}
}

func (s *UserServer) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	user, err := s.service.Register(ctx, req.GetPhoneNumber(), req.GetFullName())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidPhoneNumber), errors.Is(err, domain.ErrInvalidFullName):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, domain.ErrPhoneNumberTaken):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to register user")
		}
	}

	return &userv1.RegisterResponse{
		User: &userv1.User{
			Id:          user.ID,
			PhoneNumber: user.PhoneNumber,
			FullName:    user.FullName,
			CreatedAt:   timestamppb.New(user.CreatedAt),
		},
	}, nil
}

func (s *UserServer) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	user, err := s.service.Login(ctx, req.GetPhoneNumber())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidPhoneNumber):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, domain.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to find user")
		}
	}

	return &userv1.LoginResponse{
		User: &userv1.User{
			Id:          user.ID,
			PhoneNumber: user.PhoneNumber,
			FullName:    user.FullName,
			CreatedAt:   timestamppb.New(user.CreatedAt),
		},
	}, nil
}

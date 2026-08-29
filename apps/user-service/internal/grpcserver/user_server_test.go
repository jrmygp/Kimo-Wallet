package grpcserver

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userv1 "github.com/jrmygp/kimo-wallet/apps/user-service/gen/user/v1"
	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/domain"
)

type stubUserService struct {
	user  User
	token string
	err   error
}

// User mirrors domain.User to avoid importing it twice in the stub's field type.
type User = domain.User

func (s stubUserService) Register(ctx context.Context, phoneNumber, fullName string) (domain.User, string, error) {
	return s.user, s.token, s.err
}

func (s stubUserService) Login(ctx context.Context, phoneNumber string) (domain.User, string, error) {
	return s.user, s.token, s.err
}

func (s stubUserService) GetUserByID(ctx context.Context, kimoID string) (domain.User, error) {
	return s.user, s.err
}

func TestUserServer_Register(t *testing.T) {
	tests := []struct {
		name     string
		stub     stubUserService
		wantCode codes.Code
	}{
		{
			name: "success returns the created user",
			stub: stubUserService{
				user: domain.User{
					ID:          "11111111-1111-4111-8111-111111111111",
					PhoneNumber: "+6281234567890",
					FullName:    "Jane Doe",
					CreatedAt:   time.Unix(0, 0),
				},
				token: "signed.jwt.token",
			},
			wantCode: codes.OK,
		},
		{
			name:     "invalid phone number maps to InvalidArgument",
			stub:     stubUserService{err: domain.ErrInvalidPhoneNumber},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "invalid full name maps to InvalidArgument",
			stub:     stubUserService{err: domain.ErrInvalidFullName},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "duplicate phone number maps to AlreadyExists",
			stub:     stubUserService{err: domain.ErrPhoneNumberTaken},
			wantCode: codes.AlreadyExists,
		},
		{
			name:     "unexpected error maps to Internal and does not leak details",
			stub:     stubUserService{err: errors.New("connection refused")},
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewUserServer(tt.stub)

			resp, err := server.Register(context.Background(), &userv1.RegisterRequest{
				PhoneNumber: "+6281234567890",
				FullName:    "Jane Doe",
			})

			if tt.wantCode == codes.OK {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if resp.GetUser().GetId() != tt.stub.user.ID {
					t.Fatalf("got user id %q, want %q", resp.GetUser().GetId(), tt.stub.user.ID)
				}
				if resp.GetAccessToken() != tt.stub.token {
					t.Fatalf("got access token %q, want %q", resp.GetAccessToken(), tt.stub.token)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if got := status.Code(err); got != tt.wantCode {
				t.Fatalf("got code %v, want %v", got, tt.wantCode)
			}
			if tt.wantCode == codes.Internal && status.Convert(err).Message() == tt.stub.err.Error() {
				t.Fatalf("internal error message leaked implementation detail: %q", status.Convert(err).Message())
			}
		})
	}
}

func TestUserServer_GetUserByID(t *testing.T) {
	profilePicture := "https://example.com/avatar.png"

	tests := []struct {
		name     string
		stub     stubUserService
		wantCode codes.Code
	}{
		{
			name: "success returns the matched user, including KimoID and profile picture",
			stub: stubUserService{
				user: domain.User{
					ID:             "11111111-1111-4111-8111-111111111111",
					PhoneNumber:    "+6281234567890",
					FullName:       "Jane Doe",
					CreatedAt:      time.Unix(0, 0),
					KimoID:         "ABCDEF123456",
					ProfilePicture: &profilePicture,
				},
			},
			wantCode: codes.OK,
		},
		{
			name: "success with no profile picture set stays nil, not empty string",
			stub: stubUserService{
				user: domain.User{
					ID:          "11111111-1111-4111-8111-111111111111",
					PhoneNumber: "+6281234567890",
					FullName:    "Jane Doe",
					CreatedAt:   time.Unix(0, 0),
					KimoID:      "ABCDEF123456",
				},
			},
			wantCode: codes.OK,
		},
		{
			name:     "unknown KimoID maps to NotFound, not Internal",
			stub:     stubUserService{err: domain.ErrUserNotFound},
			wantCode: codes.NotFound,
		},
		{
			name:     "malformed KimoID maps to InvalidArgument, not Internal",
			stub:     stubUserService{err: domain.ErrInvalidKimoID},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "unexpected error maps to Internal and does not leak details",
			stub:     stubUserService{err: errors.New("connection refused")},
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewUserServer(tt.stub)

			resp, err := server.GetUserByID(context.Background(), &userv1.GetUserByIDRequest{
				KimoId: "ABCDEF123456",
			})

			if tt.wantCode == codes.OK {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if resp.GetUser().GetId() != tt.stub.user.ID {
					t.Fatalf("got user id %q, want %q", resp.GetUser().GetId(), tt.stub.user.ID)
				}
				if resp.GetUser().GetKimoId() != tt.stub.user.KimoID {
					t.Fatalf("got kimo id %q, want %q", resp.GetUser().GetKimoId(), tt.stub.user.KimoID)
				}
				gotProfilePicture := resp.GetUser().ProfilePicture
				wantProfilePicture := tt.stub.user.ProfilePicture
				switch {
				case gotProfilePicture == nil && wantProfilePicture == nil:
					// both unset, fine
				case gotProfilePicture == nil || wantProfilePicture == nil:
					t.Fatalf("got profile picture %v, want %v", gotProfilePicture, wantProfilePicture)
				case *gotProfilePicture != *wantProfilePicture:
					t.Fatalf("got profile picture %q, want %q", *gotProfilePicture, *wantProfilePicture)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if got := status.Code(err); got != tt.wantCode {
				t.Fatalf("got code %v, want %v", got, tt.wantCode)
			}
			if tt.wantCode == codes.Internal && status.Convert(err).Message() == tt.stub.err.Error() {
				t.Fatalf("internal error message leaked implementation detail: %q", status.Convert(err).Message())
			}
		})
	}
}

func TestUserServer_Login(t *testing.T) {
	tests := []struct {
		name     string
		stub     stubUserService
		wantCode codes.Code
	}{
		{
			name: "success returns the matched user",
			stub: stubUserService{
				user: domain.User{
					ID:          "11111111-1111-4111-8111-111111111111",
					PhoneNumber: "+6281234567890",
					FullName:    "Jane Doe",
					CreatedAt:   time.Unix(0, 0),
				},
				token: "signed.jwt.token",
			},
			wantCode: codes.OK,
		},
		{
			name:     "invalid phone number maps to InvalidArgument",
			stub:     stubUserService{err: domain.ErrInvalidPhoneNumber},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "unknown phone number maps to NotFound",
			stub:     stubUserService{err: domain.ErrUserNotFound},
			wantCode: codes.NotFound,
		},
		{
			name:     "unexpected error maps to Internal and does not leak details",
			stub:     stubUserService{err: errors.New("connection refused")},
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewUserServer(tt.stub)

			resp, err := server.Login(context.Background(), &userv1.LoginRequest{
				PhoneNumber: "+6281234567890",
			})

			if tt.wantCode == codes.OK {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if resp.GetUser().GetId() != tt.stub.user.ID {
					t.Fatalf("got user id %q, want %q", resp.GetUser().GetId(), tt.stub.user.ID)
				}
				if resp.GetAccessToken() != tt.stub.token {
					t.Fatalf("got access token %q, want %q", resp.GetAccessToken(), tt.stub.token)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if got := status.Code(err); got != tt.wantCode {
				t.Fatalf("got code %v, want %v", got, tt.wantCode)
			}
			if tt.wantCode == codes.Internal && status.Convert(err).Message() == tt.stub.err.Error() {
				t.Fatalf("internal error message leaked implementation detail: %q", status.Convert(err).Message())
			}
		})
	}
}

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

type stubRegistrar struct {
	user User
	err  error
}

// User mirrors domain.User to avoid importing it twice in the stub's field type.
type User = domain.User

func (s stubRegistrar) Register(ctx context.Context, phoneNumber, fullName string) (domain.User, error) {
	return s.user, s.err
}

func TestUserServer_Register(t *testing.T) {
	tests := []struct {
		name     string
		stub     stubRegistrar
		wantCode codes.Code
	}{
		{
			name: "success returns the created user",
			stub: stubRegistrar{user: domain.User{
				ID:          "11111111-1111-4111-8111-111111111111",
				PhoneNumber: "+6281234567890",
				FullName:    "Jane Doe",
				CreatedAt:   time.Unix(0, 0),
			}},
			wantCode: codes.OK,
		},
		{
			name:     "invalid phone number maps to InvalidArgument",
			stub:     stubRegistrar{err: domain.ErrInvalidPhoneNumber},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "invalid full name maps to InvalidArgument",
			stub:     stubRegistrar{err: domain.ErrInvalidFullName},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "duplicate phone number maps to AlreadyExists",
			stub:     stubRegistrar{err: domain.ErrPhoneNumberTaken},
			wantCode: codes.AlreadyExists,
		},
		{
			name:     "unexpected error maps to Internal and does not leak details",
			stub:     stubRegistrar{err: errors.New("connection refused")},
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

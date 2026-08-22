package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	userv1 "github.com/jrmygp/kimo-wallet/apps/api-gateway/gen/user/v1"
)

type stubUserServiceClient struct {
	registerResp *userv1.RegisterResponse
	registerErr  error
	loginResp    *userv1.LoginResponse
	loginErr     error
	getUserResp  *userv1.GetUserByIDResponse
	getUserErr   error
	// gotGetUserID records the id GetUserByID was actually called with, so
	// tests can assert the route param made it into the gRPC request.
	gotGetUserID *string
}

func (s stubUserServiceClient) Register(ctx context.Context, in *userv1.RegisterRequest, opts ...grpc.CallOption) (*userv1.RegisterResponse, error) {
	return s.registerResp, s.registerErr
}

func (s stubUserServiceClient) Login(ctx context.Context, in *userv1.LoginRequest, opts ...grpc.CallOption) (*userv1.LoginResponse, error) {
	return s.loginResp, s.loginErr
}

func (s stubUserServiceClient) GetUserByID(ctx context.Context, in *userv1.GetUserByIDRequest, opts ...grpc.CallOption) (*userv1.GetUserByIDResponse, error) {
	if s.gotGetUserID != nil {
		*s.gotGetUserID = in.GetId()
	}
	return s.getUserResp, s.getUserErr
}

// errorEnvelope decodes an apiResponse where data is expected to be null.
type errorEnvelope struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Data       any    `json:"data"`
}

type registerEnvelope struct {
	StatusCode int          `json:"statusCode"`
	Message    string       `json:"message"`
	Data       registerData `json:"data"`
}

type loginEnvelope struct {
	StatusCode int       `json:"statusCode"`
	Message    string    `json:"message"`
	Data       loginData `json:"data"`
}

func TestUserHandler_Register(t *testing.T) {
	fixedTime := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		body           string
		stub           stubUserServiceClient
		wantStatus     int
		wantBodyPhone  string
		wantErrPresent bool
	}{
		{
			name: "success returns 201 with the created user",
			body: `{"phoneNumber":"+6281234567890","fullName":"Jane Doe"}`,
			stub: stubUserServiceClient{registerResp: &userv1.RegisterResponse{
				User: &userv1.User{
					Id:          "11111111-1111-4111-8111-111111111111",
					PhoneNumber: "+6281234567890",
					FullName:    "Jane Doe",
					CreatedAt:   timestamppb.New(fixedTime),
				},
				AccessToken: "signed.jwt.token",
			}},
			wantStatus:    http.StatusCreated,
			wantBodyPhone: "+6281234567890",
		},
		{
			name:           "malformed JSON returns 400",
			body:           `{not json`,
			wantStatus:     http.StatusBadRequest,
			wantErrPresent: true,
		},
		{
			name:           "InvalidArgument from user-service maps to 400",
			body:           `{"phoneNumber":"bad","fullName":"Jane Doe"}`,
			stub:           stubUserServiceClient{registerErr: status.Error(codes.InvalidArgument, "phone number must be in E.164 format")},
			wantStatus:     http.StatusBadRequest,
			wantErrPresent: true,
		},
		{
			name:           "AlreadyExists from user-service maps to 409",
			body:           `{"phoneNumber":"+6281234567890","fullName":"Jane Doe"}`,
			stub:           stubUserServiceClient{registerErr: status.Error(codes.AlreadyExists, "phone number is already registered")},
			wantStatus:     http.StatusConflict,
			wantErrPresent: true,
		},
		{
			name:           "Internal error from user-service maps to 500 without leaking detail",
			body:           `{"phoneNumber":"+6281234567890","fullName":"Jane Doe"}`,
			stub:           stubUserServiceClient{registerErr: status.Error(codes.Internal, "failed to register user")},
			wantStatus:     http.StatusInternalServerError,
			wantErrPresent: true,
		},
		{
			name:           "non-gRPC error maps to 500",
			body:           `{"phoneNumber":"+6281234567890","fullName":"Jane Doe"}`,
			stub:           stubUserServiceClient{registerErr: errors.New("connection refused")},
			wantStatus:     http.StatusInternalServerError,
			wantErrPresent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUserHandler(tt.stub)

			req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			h.Register(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantErrPresent {
				var env errorEnvelope
				if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
					t.Fatalf("decode error envelope: %v", err)
				}
				if env.StatusCode != tt.wantStatus {
					t.Fatalf("envelope statusCode = %d, want %d", env.StatusCode, tt.wantStatus)
				}
				if env.Message == "" {
					t.Fatalf("expected non-empty message")
				}
				if env.Data != nil {
					t.Fatalf("expected data to be null on error, got %v", env.Data)
				}
				return
			}

			var env registerEnvelope
			if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
				t.Fatalf("decode success envelope: %v", err)
			}
			if env.StatusCode != tt.wantStatus {
				t.Fatalf("envelope statusCode = %d, want %d", env.StatusCode, tt.wantStatus)
			}
			if env.Message == "" {
				t.Fatalf("expected non-empty message")
			}
			if env.Data.User.PhoneNumber != tt.wantBodyPhone {
				t.Fatalf("got phoneNumber %q, want %q", env.Data.User.PhoneNumber, tt.wantBodyPhone)
			}
			if env.Data.User.CreatedAt != fixedTime.Format(time.RFC3339) {
				t.Fatalf("got createdAt %q, want %q", env.Data.User.CreatedAt, fixedTime.Format(time.RFC3339))
			}
			if env.Data.AccessToken != "signed.jwt.token" {
				t.Fatalf("got accessToken %q, want %q", env.Data.AccessToken, "signed.jwt.token")
			}
		})
	}
}

func TestUserHandler_Register_RejectsOversizedBody(t *testing.T) {
	h := NewUserHandler(stubUserServiceClient{})

	oversized := bytes.Repeat([]byte("a"), maxRequestBodyBytes+1)
	body := `{"phoneNumber":"+6281234567890","fullName":"` + string(oversized) + `"}`

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_Login(t *testing.T) {
	fixedTime := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		body           string
		stub           stubUserServiceClient
		wantStatus     int
		wantBodyPhone  string
		wantErrPresent bool
	}{
		{
			name: "success returns 200 with the matched user",
			body: `{"phoneNumber":"+6281234567890"}`,
			stub: stubUserServiceClient{loginResp: &userv1.LoginResponse{
				User: &userv1.User{
					Id:          "11111111-1111-4111-8111-111111111111",
					PhoneNumber: "+6281234567890",
					FullName:    "Jane Doe",
					CreatedAt:   timestamppb.New(fixedTime),
				},
				AccessToken: "signed.jwt.token",
			}},
			wantStatus:    http.StatusOK,
			wantBodyPhone: "+6281234567890",
		},
		{
			name:           "malformed JSON returns 400",
			body:           `{not json`,
			wantStatus:     http.StatusBadRequest,
			wantErrPresent: true,
		},
		{
			name:           "InvalidArgument from user-service maps to 400",
			body:           `{"phoneNumber":"bad"}`,
			stub:           stubUserServiceClient{loginErr: status.Error(codes.InvalidArgument, "phone number must be in E.164 format")},
			wantStatus:     http.StatusBadRequest,
			wantErrPresent: true,
		},
		{
			name:           "NotFound from user-service maps to 404",
			body:           `{"phoneNumber":"+6281234567890"}`,
			stub:           stubUserServiceClient{loginErr: status.Error(codes.NotFound, "user not found")},
			wantStatus:     http.StatusNotFound,
			wantErrPresent: true,
		},
		{
			name:           "Internal error from user-service maps to 500 without leaking detail",
			body:           `{"phoneNumber":"+6281234567890"}`,
			stub:           stubUserServiceClient{loginErr: status.Error(codes.Internal, "failed to find user")},
			wantStatus:     http.StatusInternalServerError,
			wantErrPresent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUserHandler(tt.stub)

			req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			h.Login(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantErrPresent {
				var env errorEnvelope
				if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
					t.Fatalf("decode error envelope: %v", err)
				}
				if env.StatusCode != tt.wantStatus {
					t.Fatalf("envelope statusCode = %d, want %d", env.StatusCode, tt.wantStatus)
				}
				if env.Message == "" {
					t.Fatalf("expected non-empty message")
				}
				if env.Data != nil {
					t.Fatalf("expected data to be null on error, got %v", env.Data)
				}
				return
			}

			var env loginEnvelope
			if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
				t.Fatalf("decode success envelope: %v", err)
			}
			if env.StatusCode != tt.wantStatus {
				t.Fatalf("envelope statusCode = %d, want %d", env.StatusCode, tt.wantStatus)
			}
			if env.Message == "" {
				t.Fatalf("expected non-empty message")
			}
			if env.Data.User.PhoneNumber != tt.wantBodyPhone {
				t.Fatalf("got phoneNumber %q, want %q", env.Data.User.PhoneNumber, tt.wantBodyPhone)
			}
			if env.Data.User.CreatedAt != fixedTime.Format(time.RFC3339) {
				t.Fatalf("got createdAt %q, want %q", env.Data.User.CreatedAt, fixedTime.Format(time.RFC3339))
			}
			if env.Data.AccessToken != "signed.jwt.token" {
				t.Fatalf("got accessToken %q, want %q", env.Data.AccessToken, "signed.jwt.token")
			}
		})
	}
}

func TestUserHandler_GetUserByID(t *testing.T) {
	fixedTime := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	const wantID = "11111111-1111-4111-8111-111111111111"

	tests := []struct {
		name           string
		stub           stubUserServiceClient
		wantStatus     int
		wantErrPresent bool
	}{
		{
			name: "success returns 200 with the requested user",
			stub: stubUserServiceClient{getUserResp: &userv1.GetUserByIDResponse{
				User: &userv1.User{
					Id:          wantID,
					PhoneNumber: "+6281234567890",
					FullName:    "Jane Doe",
					CreatedAt:   timestamppb.New(fixedTime),
				},
			}},
			wantStatus: http.StatusOK,
		},
		{
			name:           "NotFound from user-service maps to 404",
			stub:           stubUserServiceClient{getUserErr: status.Error(codes.NotFound, "user not found")},
			wantStatus:     http.StatusNotFound,
			wantErrPresent: true,
		},
		{
			name:           "Internal error from user-service maps to 500 without leaking detail",
			stub:           stubUserServiceClient{getUserErr: status.Error(codes.Internal, "failed to find user")},
			wantStatus:     http.StatusInternalServerError,
			wantErrPresent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID string
			tt.stub.gotGetUserID = &gotID
			h := NewUserHandler(tt.stub)

			req := httptest.NewRequest(http.MethodGet, "/v1/users/"+wantID, nil)
			req.SetPathValue("id", wantID)
			rec := httptest.NewRecorder()

			h.GetUserByID(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if gotID != wantID {
				t.Fatalf("GetUserByID called with id %q, want %q — route param not wired through", gotID, wantID)
			}

			if tt.wantErrPresent {
				var env errorEnvelope
				if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
					t.Fatalf("decode error envelope: %v", err)
				}
				if env.Data != nil {
					t.Fatalf("expected data to be null on error, got %v", env.Data)
				}
				return
			}

			var env struct {
				StatusCode int              `json:"statusCode"`
				Message    string           `json:"message"`
				Data       userResponseBody `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
				t.Fatalf("decode success envelope: %v", err)
			}
			if env.Data.ID != wantID {
				t.Fatalf("got id %q, want %q", env.Data.ID, wantID)
			}
			if env.Data.FullName != "Jane Doe" {
				t.Fatalf("got fullName %q, want %q", env.Data.FullName, "Jane Doe")
			}
		})
	}
}

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	userv1 "github.com/jrmygp/kimo-wallet/apps/api-gateway/gen/user/v1"
	walletv1 "github.com/jrmygp/kimo-wallet/apps/api-gateway/gen/wallet/v1"
	"github.com/jrmygp/kimo-wallet/apps/api-gateway/internal/jwtauth"
)

type stubUserServiceClient struct {
	registerResp *userv1.RegisterResponse
	registerErr  error
	loginResp    *userv1.LoginResponse
	loginErr     error
	getUserResp  *userv1.GetUserByIDResponse
	getUserErr   error
	// gotGetUserKimoID records the KimoID GetUserByID was actually called
	// with, so tests can assert the route param made it into the gRPC request.
	gotGetUserKimoID *string
}

func (s stubUserServiceClient) Register(ctx context.Context, in *userv1.RegisterRequest, opts ...grpc.CallOption) (*userv1.RegisterResponse, error) {
	return s.registerResp, s.registerErr
}

func (s stubUserServiceClient) Login(ctx context.Context, in *userv1.LoginRequest, opts ...grpc.CallOption) (*userv1.LoginResponse, error) {
	return s.loginResp, s.loginErr
}

func (s stubUserServiceClient) GetUserByID(ctx context.Context, in *userv1.GetUserByIDRequest, opts ...grpc.CallOption) (*userv1.GetUserByIDResponse, error) {
	if s.gotGetUserKimoID != nil {
		*s.gotGetUserKimoID = in.GetKimoId()
	}
	return s.getUserResp, s.getUserErr
}

type stubWalletServiceClient struct {
	getWalletResp *walletv1.GetWalletByUserIDResponse
	getWalletErr  error
	// called records whether GetWalletByUserID was invoked at all — tests
	// use this to prove the wallet client is never even called for
	// someone else's profile, not just that its result is discarded.
	called bool
}

func (s *stubWalletServiceClient) GetWalletByUserID(ctx context.Context, in *walletv1.GetWalletByUserIDRequest, opts ...grpc.CallOption) (*walletv1.GetWalletByUserIDResponse, error) {
	s.called = true
	return s.getWalletResp, s.getWalletErr
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
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
					KimoId:      "ABCDEF123456",
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
			walletStub := &stubWalletServiceClient{}
			h := NewUserHandler(tt.stub, walletStub, testLogger())

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

			// Register no longer starts a session (see Register's doc
			// comment): the response must carry neither an access token
			// nor a wallet, and the wallet client must never even be
			// called — asserted at the raw-JSON level too, not just via
			// registerData's shape, so this can't silently regress by
			// someone re-adding the field without re-adding the value.
			if walletStub.called {
				t.Fatalf("wallet client was called by Register, want it never called")
			}
			if bytes.Contains(rec.Body.Bytes(), []byte("accessToken")) {
				t.Fatalf("response contains an accessToken field, want none: %s", rec.Body.String())
			}
			if bytes.Contains(rec.Body.Bytes(), []byte("wallet")) {
				t.Fatalf("response contains a wallet field, want none: %s", rec.Body.String())
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
			if env.Data.User.KimoID != "ABCDEF123456" {
				t.Fatalf("got kimoId %q, want %q", env.Data.User.KimoID, "ABCDEF123456")
			}
			if env.Data.User.ProfilePicture != nil {
				t.Fatalf("got non-nil profilePicture %v, want nil (JSON null)", *env.Data.User.ProfilePicture)
			}
		})
	}
}

func TestUserHandler_Register_RejectsOversizedBody(t *testing.T) {
	h := NewUserHandler(stubUserServiceClient{}, &stubWalletServiceClient{}, testLogger())

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
					Id:             "11111111-1111-4111-8111-111111111111",
					PhoneNumber:    "+6281234567890",
					FullName:       "Jane Doe",
					CreatedAt:      timestamppb.New(fixedTime),
					KimoId:         "ABCDEF123456",
					ProfilePicture: proto.String("https://example.com/avatar.png"),
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
			h := NewUserHandler(tt.stub, &stubWalletServiceClient{}, testLogger())

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
			if env.Data.User.KimoID != "ABCDEF123456" {
				t.Fatalf("got kimoId %q, want %q", env.Data.User.KimoID, "ABCDEF123456")
			}
			wantProfilePicture := "https://example.com/avatar.png"
			if env.Data.User.ProfilePicture == nil || *env.Data.User.ProfilePicture != wantProfilePicture {
				t.Fatalf("got profilePicture %v, want %q", env.Data.User.ProfilePicture, wantProfilePicture)
			}
		})
	}
}

// TestUserHandler_Login_IncludesWallet proves a successful login also
// fetches and includes the caller's own wallet — unlike GetUserByID,
// there's no "someone else's balance" case to gate here: logging in is
// always the caller looking up themselves.
func TestUserHandler_Login_IncludesWallet(t *testing.T) {
	userStub := stubUserServiceClient{loginResp: &userv1.LoginResponse{
		User:        &userv1.User{Id: "11111111-1111-4111-8111-111111111111", PhoneNumber: "+6281234567890", CreatedAt: timestamppb.Now()},
		AccessToken: "signed.jwt.token",
	}}
	walletStub := &stubWalletServiceClient{getWalletResp: &walletv1.GetWalletByUserIDResponse{
		Wallet: &walletv1.Wallet{Balance: 75000, Currency: "IDR", Status: "ACTIVE"},
	}}
	h := NewUserHandler(userStub, walletStub, testLogger())

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBufferString(`{"phoneNumber":"+6281234567890"}`))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if !walletStub.called {
		t.Fatalf("wallet client was never called on a successful login")
	}

	var env loginEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if env.Data.Wallet == nil {
		t.Fatalf("got nil wallet on a successful login, want it populated")
	}
	if env.Data.Wallet.Balance != 75000 || env.Data.Wallet.Currency != "IDR" || env.Data.Wallet.Status != "ACTIVE" {
		t.Fatalf("got wallet %+v, want {Balance:75000 Currency:IDR Status:ACTIVE}", env.Data.Wallet)
	}
}

// TestUserHandler_Login_WalletLookupFailureDoesNotFailLogin is the bug
// this was actually fixed for: a wallet-lookup failure must never break
// login itself — the user's credentials were valid and user-service
// already authenticated them successfully.
func TestUserHandler_Login_WalletLookupFailureDoesNotFailLogin(t *testing.T) {
	userStub := stubUserServiceClient{loginResp: &userv1.LoginResponse{
		User:        &userv1.User{Id: "11111111-1111-4111-8111-111111111111", PhoneNumber: "+6281234567890", CreatedAt: timestamppb.Now()},
		AccessToken: "signed.jwt.token",
	}}
	walletStub := &stubWalletServiceClient{getWalletErr: status.Error(codes.NotFound, "wallet not found")}
	h := NewUserHandler(userStub, walletStub, testLogger())

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBufferString(`{"phoneNumber":"+6281234567890"}`))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s) — a wallet lookup failure must not fail login", rec.Code, http.StatusOK, rec.Body.String())
	}

	var env loginEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if env.Data.AccessToken != "signed.jwt.token" {
		t.Fatalf("got accessToken %q, want %q — login itself must still succeed", env.Data.AccessToken, "signed.jwt.token")
	}
	if env.Data.Wallet != nil {
		t.Fatalf("got non-nil wallet %+v after a failed wallet lookup, want nil", env.Data.Wallet)
	}
}

func TestUserHandler_GetUserByID(t *testing.T) {
	fixedTime := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	const wantID = "11111111-1111-4111-8111-111111111111"
	const wantKimoID = "ABCDEF123456"

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
					KimoId:      wantKimoID,
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
			var gotKimoID string
			tt.stub.gotGetUserKimoID = &gotKimoID
			walletStub := &stubWalletServiceClient{}
			h := NewUserHandler(tt.stub, walletStub, testLogger())

			// No jwtauth claims on this request's context at all — same as
			// every request in this table. Proves the no-claims case
			// (e.g. some future unauthenticated caller of this handler,
			// or a test that forgot to set them up) degrades to "no
			// wallet", not a panic or a leaked wallet.
			req := httptest.NewRequest(http.MethodGet, "/v1/user/"+wantKimoID, nil)
			req.SetPathValue("kimoId", wantKimoID)
			rec := httptest.NewRecorder()

			h.GetUserByID(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if gotKimoID != wantKimoID {
				t.Fatalf("GetUserByID called with kimoId %q, want %q — route param not wired through", gotKimoID, wantKimoID)
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
				StatusCode int      `json:"statusCode"`
				Message    string   `json:"message"`
				Data       userData `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
				t.Fatalf("decode success envelope: %v", err)
			}
			if env.Data.User.ID != wantID {
				t.Fatalf("got id %q, want %q", env.Data.User.ID, wantID)
			}
			if env.Data.User.FullName != "Jane Doe" {
				t.Fatalf("got fullName %q, want %q", env.Data.User.FullName, "Jane Doe")
			}
			if env.Data.User.KimoID != wantKimoID {
				t.Fatalf("got kimoId %q, want %q", env.Data.User.KimoID, wantKimoID)
			}
			if env.Data.User.ProfilePicture != nil {
				t.Fatalf("got non-nil profilePicture %v, want nil (JSON null)", *env.Data.User.ProfilePicture)
			}
			if env.Data.Wallet != nil {
				t.Fatalf("got non-nil wallet %+v with no auth claims on the request, want nil", env.Data.Wallet)
			}
			if walletStub.called {
				t.Fatalf("wallet client was called with no auth claims on the request — must never even attempt a wallet lookup without a caller identity")
			}
		})
	}
}

// withClaims returns a copy of req carrying jwtauth claims for subject —
// simulates what middleware.RequireAuth already put on the context by
// the time a real request reaches this handler.
func withClaims(req *http.Request, subject string) *http.Request {
	claims := &jwtauth.Claims{}
	claims.Subject = subject
	return req.WithContext(jwtauth.WithClaims(req.Context(), claims))
}

// TestUserHandler_GetUserByID_IncludesOwnWallet proves the one case
// wallet enrichment is actually supposed to fire: the authenticated
// caller looking up their own profile.
func TestUserHandler_GetUserByID_IncludesOwnWallet(t *testing.T) {
	const selfID = "11111111-1111-4111-8111-111111111111"
	const selfKimoID = "ABCDEF123456"

	userStub := stubUserServiceClient{getUserResp: &userv1.GetUserByIDResponse{
		User: &userv1.User{Id: selfID, FullName: "Jane Doe", KimoId: selfKimoID, CreatedAt: timestamppb.Now()},
	}}
	walletStub := &stubWalletServiceClient{getWalletResp: &walletv1.GetWalletByUserIDResponse{
		Wallet: &walletv1.Wallet{Balance: 50000, Currency: "IDR", Status: "ACTIVE"},
	}}
	h := NewUserHandler(userStub, walletStub, testLogger())

	req := withClaims(httptest.NewRequest(http.MethodGet, "/v1/user/"+selfKimoID, nil), selfID)
	req.SetPathValue("kimoId", selfKimoID)
	rec := httptest.NewRecorder()

	h.GetUserByID(rec, req)

	if !walletStub.called {
		t.Fatalf("wallet client was never called for the caller looking up their own profile")
	}

	var env struct {
		Data userData `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if env.Data.Wallet == nil {
		t.Fatalf("got nil wallet for the caller's own profile, want it populated")
	}
	if env.Data.Wallet.Balance != 50000 || env.Data.Wallet.Currency != "IDR" || env.Data.Wallet.Status != "ACTIVE" {
		t.Fatalf("got wallet %+v, want {Balance:50000 Currency:IDR Status:ACTIVE}", env.Data.Wallet)
	}
}

// TestUserHandler_GetUserByID_OmitsWalletForOtherUsers is the actual
// privacy guarantee: searching someone else's KimoID must never even
// attempt to fetch their wallet, let alone return it.
func TestUserHandler_GetUserByID_OmitsWalletForOtherUsers(t *testing.T) {
	const otherUserID = "11111111-1111-4111-8111-111111111111"
	const callerID = "22222222-2222-4222-8222-222222222222"
	const otherKimoID = "ABCDEF123456"

	userStub := stubUserServiceClient{getUserResp: &userv1.GetUserByIDResponse{
		User: &userv1.User{Id: otherUserID, FullName: "Jane Doe", KimoId: otherKimoID, CreatedAt: timestamppb.Now()},
	}}
	walletStub := &stubWalletServiceClient{getWalletResp: &walletv1.GetWalletByUserIDResponse{
		Wallet: &walletv1.Wallet{Balance: 999999999, Currency: "IDR", Status: "ACTIVE"},
	}}
	h := NewUserHandler(userStub, walletStub, testLogger())

	req := withClaims(httptest.NewRequest(http.MethodGet, "/v1/user/"+otherKimoID, nil), callerID)
	req.SetPathValue("kimoId", otherKimoID)
	rec := httptest.NewRecorder()

	h.GetUserByID(rec, req)

	if walletStub.called {
		t.Fatalf("wallet client was called for a user other than the caller — this must never happen, it's a balance-privacy leak")
	}

	var env struct {
		Data userData `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if env.Data.Wallet != nil {
		t.Fatalf("got non-nil wallet %+v for another user's profile, want nil", env.Data.Wallet)
	}
}

// TestUserHandler_GetUserByID_WalletLookupFailureDoesNotFailRequest
// proves wallet enrichment is best-effort: a wallet lookup failure (e.g.
// not provisioned yet — wallet creation is async) must not fail the
// whole request, since the user profile itself was already found fine.
func TestUserHandler_GetUserByID_WalletLookupFailureDoesNotFailRequest(t *testing.T) {
	const selfID = "11111111-1111-4111-8111-111111111111"
	const selfKimoID = "ABCDEF123456"

	userStub := stubUserServiceClient{getUserResp: &userv1.GetUserByIDResponse{
		User: &userv1.User{Id: selfID, FullName: "Jane Doe", KimoId: selfKimoID, CreatedAt: timestamppb.Now()},
	}}
	walletStub := &stubWalletServiceClient{getWalletErr: status.Error(codes.NotFound, "wallet not found")}
	h := NewUserHandler(userStub, walletStub, testLogger())

	req := withClaims(httptest.NewRequest(http.MethodGet, "/v1/user/"+selfKimoID, nil), selfID)
	req.SetPathValue("kimoId", selfKimoID)
	rec := httptest.NewRecorder()

	h.GetUserByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d — a wallet lookup failure must not fail the whole request", rec.Code, http.StatusOK)
	}

	var env struct {
		Data userData `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if env.Data.User.ID != selfID {
		t.Fatalf("got user id %q, want %q — user data must still be present", env.Data.User.ID, selfID)
	}
	if env.Data.Wallet != nil {
		t.Fatalf("got non-nil wallet %+v after a failed wallet lookup, want nil", env.Data.Wallet)
	}
}

func TestUserHandler_GetUserByID_MissingID(t *testing.T) {
	h := NewUserHandler(stubUserServiceClient{}, &stubWalletServiceClient{}, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/v1/user/", nil)
	req.SetPathValue("kimoId", "")
	rec := httptest.NewRecorder()

	h.GetUserByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

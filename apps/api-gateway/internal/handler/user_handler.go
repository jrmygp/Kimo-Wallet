// Package handler translates HTTP/JSON requests into gRPC calls against
// internal services and translates the gRPC response/error back to
// HTTP/JSON. Per docs/CLAUDE.md §4, the gateway holds no domain logic —
// validation of *business* rules (phone number format, name length,
// uniqueness) happens in the owning service, not here.
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"google.golang.org/grpc"

	userv1 "github.com/jrmygp/kimo-wallet/apps/api-gateway/gen/user/v1"
)

// maxRequestBodyBytes bounds request bodies read by this handler — a
// boundary-input guard (docs/CLAUDE.md §5.1 "validate at the boundary"),
// not a business rule.
const maxRequestBodyBytes = 1 << 20 // 1MiB

// userServiceClient is the subset of userv1.UserServiceClient this handler
// depends on, so it can be exercised in tests with a stub instead of a
// real gRPC connection.
type userServiceClient interface {
	Register(ctx context.Context, in *userv1.RegisterRequest, opts ...grpc.CallOption) (*userv1.RegisterResponse, error)
	Login(ctx context.Context, in *userv1.LoginRequest, opts ...grpc.CallOption) (*userv1.LoginResponse, error)
	GetUserByID(ctx context.Context, in *userv1.GetUserByIDRequest, opts ...grpc.CallOption) (*userv1.GetUserByIDResponse, error)
}

type UserHandler struct {
	client userServiceClient
}

func NewUserHandler(client userServiceClient) *UserHandler {
	return &UserHandler{client: client}
}

type registerRequestBody struct {
	PhoneNumber string `json:"phoneNumber"`
	FullName    string `json:"fullName"`
}

type userResponseBody struct {
	ID          string `json:"id"`
	PhoneNumber string `json:"phoneNumber"`
	FullName    string `json:"fullName"`
	CreatedAt   string `json:"createdAt"`
}

type registerData struct {
	User        userResponseBody `json:"user"`
	AccessToken string           `json:"accessToken"`
}

type loginRequestBody struct {
	PhoneNumber string `json:"phoneNumber"`
}

type loginData struct {
	User        userResponseBody `json:"user"`
	AccessToken string           `json:"accessToken"`
}

type userData struct {
	User userResponseBody `json:"user"`
}

// Register handles POST /v1/auth/register. It only decodes the request
// shape and maps the result; phone number / full name validation is
// user-service's job, not this handler's.
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	var body registerRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	resp, err := h.client.Register(r.Context(), &userv1.RegisterRequest{
		PhoneNumber: body.PhoneNumber,
		FullName:    body.FullName,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, "user registered successfully", registerData{
		User: userResponseBody{
			ID:          resp.GetUser().GetId(),
			PhoneNumber: resp.GetUser().GetPhoneNumber(),
			FullName:    resp.GetUser().GetFullName(),
			CreatedAt:   resp.GetUser().GetCreatedAt().AsTime().Format(time.RFC3339),
		},
		AccessToken: resp.GetAccessToken(),
	})
}

// Login handles POST /v1/auth/login. Same shape as Register: decode,
// delegate, translate — no business logic (§4) belongs here. Whether the
// phone number is well-formed or actually exists is user-service's call,
// surfaced back to the client via writeGRPCError.
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	var body loginRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	resp, err := h.client.Login(r.Context(), &userv1.LoginRequest{
		PhoneNumber: body.PhoneNumber,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "login successful", loginData{
		User: userResponseBody{
			ID:          resp.GetUser().GetId(),
			PhoneNumber: resp.GetUser().GetPhoneNumber(),
			FullName:    resp.GetUser().GetFullName(),
			CreatedAt:   resp.GetUser().GetCreatedAt().AsTime().Format(time.RFC3339),
		},
		AccessToken: resp.GetAccessToken(),
	})
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	// How do i retrieve the id from route param?
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "ID is invalid")
		return
	}

	resp, err := h.client.GetUserByID(r.Context(), &userv1.GetUserByIDRequest{
		Id: id,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "ok", userData{
		User: userResponseBody{
			ID:          resp.GetUser().GetId(),
			PhoneNumber: resp.GetUser().GetPhoneNumber(),
			FullName:    resp.GetUser().GetFullName(),
			CreatedAt:   resp.GetUser().GetCreatedAt().AsTime().Format(time.RFC3339),
		},
	})
}

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

// userRegisterer is the subset of userv1.UserServiceClient this handler
// depends on, so it can be exercised in tests with a stub instead of a
// real gRPC connection.
type userRegisterer interface {
	Register(ctx context.Context, in *userv1.RegisterRequest, opts ...grpc.CallOption) (*userv1.RegisterResponse, error)
}

type UserHandler struct {
	client userRegisterer
}

func NewUserHandler(client userRegisterer) *UserHandler {
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

type registerResponseBody struct {
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

	writeJSON(w, http.StatusCreated, registerResponseBody{
		User: userResponseBody{
			ID:          resp.GetUser().GetId(),
			PhoneNumber: resp.GetUser().GetPhoneNumber(),
			FullName:    resp.GetUser().GetFullName(),
			CreatedAt:   resp.GetUser().GetCreatedAt().AsTime().Format(time.RFC3339),
		},
	})
}

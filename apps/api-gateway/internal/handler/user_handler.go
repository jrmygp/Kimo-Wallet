// Package handler translates HTTP/JSON requests into gRPC calls against
// internal services and translates the gRPC response/error back to
// HTTP/JSON. Per docs/CLAUDE.md §4, the gateway holds no domain logic —
// validation of *business* rules (phone number format, name length,
// uniqueness) happens in the owning service, not here.
package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"google.golang.org/grpc"

	userv1 "github.com/jrmygp/kimo-wallet/apps/api-gateway/gen/user/v1"
	walletv1 "github.com/jrmygp/kimo-wallet/apps/api-gateway/gen/wallet/v1"
	"github.com/jrmygp/kimo-wallet/apps/api-gateway/internal/jwtauth"
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

// walletServiceClient is the subset of walletv1.WalletServiceClient this
// handler depends on — used only to enrich GetUserByID's response with
// the caller's own wallet (see GetUserByID: never called, let alone
// exposed, for anyone else's wallet).
type walletServiceClient interface {
	GetWalletByUserID(ctx context.Context, in *walletv1.GetWalletByUserIDRequest, opts ...grpc.CallOption) (*walletv1.GetWalletByUserIDResponse, error)
}

type UserHandler struct {
	client       userServiceClient
	walletClient walletServiceClient
	logger       *slog.Logger
}

func NewUserHandler(client userServiceClient, walletClient walletServiceClient, logger *slog.Logger) *UserHandler {
	return &UserHandler{client: client, walletClient: walletClient, logger: logger}
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
	// ProfilePicture is a pointer, not a plain string, so an unset picture
	// serializes as JSON null rather than "" — the two are not the same
	// thing to a client.
	ProfilePicture *string `json:"profilePicture"`
	KimoID         string  `json:"kimoId"`
}

// toUserResponseBody maps the gRPC User message to this handler's JSON
// shape — shared by Register/Login/GetUserByID's responses.
func toUserResponseBody(user *userv1.User) userResponseBody {
	return userResponseBody{
		ID:             user.GetId(),
		PhoneNumber:    user.GetPhoneNumber(),
		FullName:       user.GetFullName(),
		CreatedAt:      user.GetCreatedAt().AsTime().Format(time.RFC3339),
		ProfilePicture: user.ProfilePicture,
		KimoID:         user.GetKimoId(),
	}
}

// registerData deliberately carries neither an access token nor a wallet:
// registering no longer starts a session — see Register's doc comment.
type registerData struct {
	User userResponseBody `json:"user"`
}

type loginRequestBody struct {
	PhoneNumber string `json:"phoneNumber"`
}

type loginData struct {
	User userResponseBody `json:"user"`
	// Wallet is never nil-because-of-privacy the way GetUserByID's is:
	// logging in is always the caller looking up *themselves*, so there's
	// no "someone else's balance" case here — see fetchWallet. It's still
	// a pointer because the lookup itself is still best-effort (see
	// fetchWallet's doc comment).
	Wallet      *walletResponseBody `json:"wallet"`
	AccessToken string              `json:"accessToken"`
}

// walletResponseBody is deliberately only ever populated on
// GetUserByID's response when the authenticated caller is looking up
// their own profile — see GetUserByID. Balance is a raw JSON integer
// (int64), not a string: docs/CLAUDE.md §3.3 rule 4 permits either for
// money serialization, just never a float.
type walletResponseBody struct {
	Balance  int64  `json:"balance"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
}

type userData struct {
	User userResponseBody `json:"user"`
	// Wallet is nil (JSON null) unless the authenticated caller's own id
	// matches the looked-up user's — see GetUserByID. This endpoint is
	// primarily used to look up *other* users (Transfer search), and a
	// wallet balance is private to its owner.
	Wallet *walletResponseBody `json:"wallet"`
}

// Register handles POST /v1/auth/register. It only decodes the request
// shape and maps the result; phone number / full name validation is
// user-service's job, not this handler's.
//
// Deliberately does not authenticate the caller: registering creates the
// account but does not start a session — the client must call Login
// separately with the same phone number. This is why the response carries
// neither an access token nor a wallet (unlike Login/GetUserByID): going
// through a separate Login step also means, in practice, that the async
// Kafka wallet-provisioning race (see fetchWallet's doc comment) has
// usually already resolved by the time the user is authenticated at all.
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
		User: toUserResponseBody(resp.GetUser()),
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
		User:        toUserResponseBody(resp.GetUser()),
		Wallet:      h.fetchWallet(r.Context(), resp.GetUser().GetId()),
		AccessToken: resp.GetAccessToken(),
	})
}

// GetUserByID handles GET /v1/user/{kimoId}. Despite the RPC's name, the
// path segment is the user's KimoID (see packages/contracts/user/v1/user.proto),
// not the internal `id` — kept unrenamed to match the RPC/handler name.
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	kimoID := r.PathValue("kimoId")
	if kimoID == "" {
		writeError(w, http.StatusBadRequest, "KimoID is invalid")
		return
	}

	resp, err := h.client.GetUserByID(r.Context(), &userv1.GetUserByIDRequest{
		KimoId: kimoID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "ok", userData{
		User:   toUserResponseBody(resp.GetUser()),
		Wallet: h.walletForOwnProfile(r.Context(), resp.GetUser().GetId()),
	})
}

// walletForOwnProfile returns lookedUpUserID's wallet, but only if the
// authenticated caller (from the JWT claims RequireAuth put on the
// context) *is* lookedUpUserID — this handler is primarily used to look
// up other users (Transfer search), and a wallet balance is private to
// its owner, so it's never even fetched, let alone returned, for anyone
// else's id. Returns nil if the caller doesn't match or there are no
// claims on the context at all (e.g. GetUserByID is the only caller of
// this, and it's always authenticated, but this stays defensive rather
// than assuming that never changes); see fetchWallet for the rest.
func (h *UserHandler) walletForOwnProfile(ctx context.Context, lookedUpUserID string) *walletResponseBody {
	claims, ok := jwtauth.ClaimsFromContext(ctx)
	if !ok || claims.Subject != lookedUpUserID {
		return nil
	}

	return h.fetchWallet(ctx, claims.Subject)
}

// fetchWallet fetches userID's wallet and converts it to the JSON shape,
// or returns nil (never an error) if the lookup fails. Wallet enrichment
// is always best-effort on top of an already-successful user operation
// (login or profile lookup) — never something that fails the request
// itself. This matters especially right after registration/login: wallet
// provisioning is asynchronous (Kafka — see 2026-09-05's entries), so a
// brand new user briefly having no wallet yet is an expected race, not
// an error, and must never turn a successful login into a broken one.
func (h *UserHandler) fetchWallet(ctx context.Context, userID string) *walletResponseBody {
	resp, err := h.walletClient.GetWalletByUserID(ctx, &walletv1.GetWalletByUserIDRequest{
		UserId: userID,
	})
	if err != nil {
		h.logger.Error("fetch wallet", "error", err.Error(), "user_id", userID)
		return nil
	}

	return &walletResponseBody{
		Balance:  resp.GetWallet().GetBalance(),
		Currency: resp.GetWallet().GetCurrency(),
		Status:   resp.GetWallet().GetStatus(),
	}
}

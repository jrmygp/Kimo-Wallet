package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	transactionv1 "github.com/jrmygp/kimo-wallet/apps/api-gateway/gen/transaction/v1"
	"github.com/jrmygp/kimo-wallet/apps/api-gateway/internal/jwtauth"
	"google.golang.org/grpc"
)

type transactionServiceClient interface {
	CreateTransaction(ctx context.Context, in *transactionv1.CreateTransactionRequest, opts ...grpc.CallOption) (*transactionv1.CreateTransactionResponse, error)
}

type TransactionHandler struct {
	client transactionServiceClient
	logger *slog.Logger
}

func NewTransactionHandler(client transactionServiceClient, logger *slog.Logger) *TransactionHandler {
	return &TransactionHandler{client: client, logger: logger}
}

type createTransactionRequestBody struct {
	IdempotencyKey   string `json:"idempotencyKey"`
	SenderUserID     string `json:"senderUserId"`
	SenderWalletID   string `json:"senderWalletId"`
	ReceiverUserID   string `json:"receiverUserId"`
	ReceiverWalletID string `json:"receiverWalletId"`
	Currency         string `json:"currency"`
	Amount           int64  `json:"amount"`
}

type createTransactionResponseBody struct {
	ID               string `json:"id"`
	IdempotencyKey   string `json:"idempotencyKey"`
	SenderUserID     string `json:"senderUserId"`
	SenderWalletID   string `json:"senderWalletId"`
	ReceiverUserID   string `json:"receiverUserId"`
	ReceiverWalletID string `json:"receiverWalletId"`
	Currency         string `json:"currency"`
	Amount           int64  `json:"amount"`
	Status           string `json:"status"`
	CreatedAt        string `json:"createdAt"`
}

type transactionData struct {
	Transaction createTransactionResponseBody `json:"transaction"`
}

func convertTransactionToResponseBody(t *transactionv1.Transaction) createTransactionResponseBody {
	return createTransactionResponseBody{
		ID:               t.Id,
		IdempotencyKey:   t.IdempotencyKey,
		SenderUserID:     t.SenderUserId,
		SenderWalletID:   t.SenderWalletId,
		ReceiverUserID:   t.ReceiverUserId,
		ReceiverWalletID: t.ReceiverWalletId,
		Currency:         t.Currency,
		Amount:           t.Amount,
		Status:           t.Status,
		CreatedAt:        t.CreatedAt.AsTime().Format(time.RFC3339),
	}
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	var body createTransactionRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// The gateway holds no domain logic (docs/CLAUDE.md §4), but
	// authorizing the request against the resource owner is not domain
	// logic — it's the one thing only the gateway can do, since it's the
	// only layer that sees the caller's own authenticated identity
	// alongside the sender id the client put in the body. Without this,
	// any authenticated user could move money out of any wallet just by
	// putting someone else's id here (§6).
	claims, ok := jwtauth.ClaimsFromContext(r.Context())
	if !ok {
		// RequireAuth always puts claims on the context before this
		// handler runs — reaching here without them means this handler
		// was wired up without that middleware, not a normal
		// unauthenticated request (those are already rejected with 401
		// by RequireAuth itself, before this handler is ever called).
		h.logger.Error("create transaction: no claims on request context — handler wired without RequireAuth?")
		writeError(w, http.StatusForbidden, "sender must be the authenticated user")
		return
	}
	if claims.Subject != body.SenderUserID {
		h.logger.Error("create transaction: sender does not match authenticated caller",
			"authenticated_user_id", claims.Subject, "requested_sender_user_id", body.SenderUserID)
		writeError(w, http.StatusForbidden, "sender must be the authenticated user")
		return
	}

	resp, err := h.client.CreateTransaction(r.Context(), &transactionv1.CreateTransactionRequest{
		IdempotencyKey:   body.IdempotencyKey,
		SenderUserId:     body.SenderUserID,
		SenderWalletId:   body.SenderWalletID,
		ReceiverUserId:   body.ReceiverUserID,
		ReceiverWalletId: body.ReceiverWalletID,
		Currency:         body.Currency,
		Amount:           body.Amount,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, "transaction created successfully", transactionData{
		Transaction: convertTransactionToResponseBody(resp.GetTransaction()),
	})
}

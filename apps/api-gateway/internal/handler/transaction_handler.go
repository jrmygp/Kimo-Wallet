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

// createTransactionRequestBody deliberately has no sender and no wallet ids:
// the sender is always the authenticated caller (taken from the JWT below),
// and both wallets are resolved by transaction-service from the user ids.
// A client that still sends senderUserId/senderWalletId/receiverWalletId has
// them ignored by the JSON decoder, never trusted.
type createTransactionRequestBody struct {
	IdempotencyKey string  `json:"idempotencyKey"`
	ReceiverUserID string  `json:"receiverUserId"`
	Currency       string  `json:"currency"`
	Amount         int64   `json:"amount"`
	Note           *string `json:"note"`
}

// createTransactionResponseBody omits wallet ids: the client never deals in
// them, and the receiver's wallet id is no business of the sender's.
type createTransactionResponseBody struct {
	ID             string `json:"id"`
	IdempotencyKey string `json:"idempotencyKey"`
	SenderUserID   string `json:"senderUserId"`
	ReceiverUserID string `json:"receiverUserId"`
	Currency       string `json:"currency"`
	Amount         int64  `json:"amount"`
	Status         string `json:"status"`
	Note           string `json:"note"`
	CreatedAt      string `json:"createdAt"`
}

type transactionData struct {
	Transaction createTransactionResponseBody `json:"transaction"`
}

func convertTransactionToResponseBody(t *transactionv1.Transaction) createTransactionResponseBody {
	return createTransactionResponseBody{
		ID:             t.Id,
		IdempotencyKey: t.IdempotencyKey,
		SenderUserID:   t.SenderUserId,
		ReceiverUserID: t.ReceiverUserId,
		Currency:       t.Currency,
		Amount:         t.Amount,
		Status:         t.Status,
		Note:           t.GetNote(),
		CreatedAt:      t.CreatedAt.AsTime().Format(time.RFC3339),
	}
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	var body createTransactionRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// The sender is the authenticated caller — taken from the JWT, never
	// from the body. Identifying the caller is not domain logic (docs/CLAUDE.md
	// §4); it's the one thing only the gateway can do, since it's the only
	// layer that sees the verified token. Without this, any authenticated user
	// could move money out of any wallet by naming someone else as sender (§6).
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

	// body.Note is nil whenever the client omits it — that's the whole
	// point of it being optional, so it's converted to the request's plain
	// (always-present) string field rather than dereferenced directly.
	var note string
	if body.Note != nil {
		note = *body.Note
	}

	resp, err := h.client.CreateTransaction(r.Context(), &transactionv1.CreateTransactionRequest{
		IdempotencyKey: body.IdempotencyKey,
		SenderUserId:   claims.Subject,
		ReceiverUserId: body.ReceiverUserID,
		Currency:       body.Currency,
		Amount:         body.Amount,
		Note:           note,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, "transaction created successfully", transactionData{
		Transaction: convertTransactionToResponseBody(resp.GetTransaction()),
	})
}

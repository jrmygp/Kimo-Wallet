package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	transactionv1 "github.com/jrmygp/kimo-wallet/apps/api-gateway/gen/transaction/v1"
	"google.golang.org/grpc"
)

type transactionServiceClient interface {
	CreateTransaction(ctx context.Context, in *transactionv1.CreateTransactionRequest, opts ...grpc.CallOption) (*transactionv1.CreateTransactionResponse, error)
}

type TransactionHandler struct {
	client transactionServiceClient
}

func NewTransactionHandler(client transactionServiceClient, logger *slog.Logger) *TransactionHandler {
	return &TransactionHandler{client: client}
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
		CreatedAt:        t.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	var body createTransactionRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
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

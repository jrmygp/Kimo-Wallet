package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	transactionv1 "github.com/jrmygp/kimo-wallet/apps/api-gateway/gen/transaction/v1"
)

const (
	callerUserID   = "11111111-1111-4111-8111-111111111111"
	receiverUserID = "33333333-3333-4333-8333-333333333333"
)

type stubTransactionServiceClient struct {
	resp *transactionv1.CreateTransactionResponse
	err  error
	// got records the request the handler actually sent, so tests can assert
	// what reached transaction-service — not just the HTTP status.
	got *transactionv1.CreateTransactionRequest
}

func (s *stubTransactionServiceClient) CreateTransaction(ctx context.Context, in *transactionv1.CreateTransactionRequest, opts ...grpc.CallOption) (*transactionv1.CreateTransactionResponse, error) {
	s.got = in
	return s.resp, s.err
}

func pendingTransactionResponse() *transactionv1.CreateTransactionResponse {
	return &transactionv1.CreateTransactionResponse{Transaction: &transactionv1.Transaction{
		Id:               "txn-1",
		IdempotencyKey:   "intent-123",
		SenderUserId:     callerUserID,
		SenderWalletId:   "22222222-2222-4222-8222-222222222222",
		ReceiverUserId:   receiverUserID,
		ReceiverWalletId: "44444444-4444-4444-8444-444444444444",
		Amount:           50000,
		Currency:         "IDR",
		Status:           "PENDING",
		CreatedAt:        timestamppb.New(time.Date(2026, 10, 4, 9, 0, 0, 0, time.UTC)),
	}}
}

func newCreateTransactionRequest(body string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/v1/transaction", bytes.NewBufferString(body))
}

// TestTransactionHandler_CreateTransaction_SenderComesFromJWT proves the
// sender sent to transaction-service is the authenticated caller, even when
// the body names someone else as sender and supplies wallet ids — those
// fields are not part of the request contract and must never be forwarded
// (docs/CLAUDE.md §6).
func TestTransactionHandler_CreateTransaction_SenderComesFromJWT(t *testing.T) {
	client := &stubTransactionServiceClient{resp: pendingTransactionResponse()}
	h := NewTransactionHandler(client, testLogger())

	body := `{
		"idempotencyKey": "intent-123",
		"senderUserId": "99999999-9999-4999-8999-999999999999",
		"senderWalletId": "88888888-8888-4888-8888-888888888888",
		"receiverUserId": "` + receiverUserID + `",
		"receiverWalletId": "77777777-7777-4777-8777-777777777777",
		"currency": "IDR",
		"amount": 50000
	}`
	rec := httptest.NewRecorder()
	h.CreateTransaction(rec, withClaims(newCreateTransactionRequest(body), callerUserID))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if client.got == nil {
		t.Fatalf("transaction-service was never called")
	}
	if client.got.GetSenderUserId() != callerUserID {
		t.Fatalf("forwarded SenderUserId = %q, want the authenticated caller %q", client.got.GetSenderUserId(), callerUserID)
	}
	if client.got.GetReceiverUserId() != receiverUserID {
		t.Fatalf("forwarded ReceiverUserId = %q, want %q", client.got.GetReceiverUserId(), receiverUserID)
	}
	if client.got.GetIdempotencyKey() != "intent-123" || client.got.GetAmount() != 50000 || client.got.GetCurrency() != "IDR" {
		t.Fatalf("forwarded request = %+v, want idempotency key, amount and currency passed through unchanged", client.got)
	}
}

// TestTransactionHandler_CreateTransaction_ResponseHasNoWalletIDs proves the
// HTTP response never exposes wallet ids — in particular not the receiver's.
func TestTransactionHandler_CreateTransaction_ResponseHasNoWalletIDs(t *testing.T) {
	client := &stubTransactionServiceClient{resp: pendingTransactionResponse()}
	h := NewTransactionHandler(client, testLogger())

	body := `{"idempotencyKey":"intent-123","receiverUserId":"` + receiverUserID + `","currency":"IDR","amount":50000}`
	rec := httptest.NewRecorder()
	h.CreateTransaction(rec, withClaims(newCreateTransactionRequest(body), callerUserID))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "WalletId") || strings.Contains(rec.Body.String(), "walletId") {
		t.Fatalf("response leaks a wallet id: %s", rec.Body.String())
	}

	var envelope struct {
		Data transactionData `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Data.Transaction.ID != "txn-1" || envelope.Data.Transaction.Status != "PENDING" {
		t.Fatalf("response transaction = %+v, want id txn-1 with status PENDING", envelope.Data.Transaction)
	}
}

// TestTransactionHandler_CreateTransaction_NoClaimsIsForbidden proves a
// request that somehow reaches the handler without RequireAuth's claims is
// rejected before transaction-service is called — there is no sender to use.
func TestTransactionHandler_CreateTransaction_NoClaimsIsForbidden(t *testing.T) {
	client := &stubTransactionServiceClient{resp: pendingTransactionResponse()}
	h := NewTransactionHandler(client, testLogger())

	body := `{"idempotencyKey":"intent-123","receiverUserId":"` + receiverUserID + `","currency":"IDR","amount":50000}`
	rec := httptest.NewRecorder()
	h.CreateTransaction(rec, newCreateTransactionRequest(body))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if client.got != nil {
		t.Fatalf("transaction-service was called without an authenticated sender")
	}
}

func TestTransactionHandler_CreateTransaction_ErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		err        error
		wantStatus int
	}{
		{
			name:       "malformed JSON returns 400 without calling transaction-service",
			body:       `{not json`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "InvalidArgument maps to 400",
			body:       `{"idempotencyKey":"intent-123","receiverUserId":"` + callerUserID + `","currency":"IDR","amount":50000}`,
			err:        status.Error(codes.InvalidArgument, "sender and receiver must be different users"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "NotFound (receiver has no wallet) maps to 404",
			body:       `{"idempotencyKey":"intent-123","receiverUserId":"` + receiverUserID + `","currency":"IDR","amount":50000}`,
			err:        status.Error(codes.NotFound, "receiver user not found"),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "AlreadyExists (repeated idempotency key) maps to 409",
			body:       `{"idempotencyKey":"intent-123","receiverUserId":"` + receiverUserID + `","currency":"IDR","amount":50000}`,
			err:        status.Error(codes.AlreadyExists, "idempotency key conflict"),
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &stubTransactionServiceClient{err: tt.err}
			h := NewTransactionHandler(client, testLogger())

			rec := httptest.NewRecorder()
			h.CreateTransaction(rec, withClaims(newCreateTransactionRequest(tt.body), callerUserID))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.err == nil && client.got != nil {
				t.Fatalf("transaction-service was called for a request that should have been rejected at the gateway")
			}
		})
	}
}

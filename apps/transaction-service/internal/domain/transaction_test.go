package domain

import (
	"errors"
	"testing"
)

func TestNewTransactionRequest(t *testing.T) {
	const (
		senderUserID     = "11111111-1111-4111-8111-111111111111"
		senderWalletID   = "22222222-2222-4222-8222-222222222222"
		receiverUserID   = "33333333-3333-4333-8333-333333333333"
		receiverWalletID = "44444444-4444-4444-8444-444444444444"
	)

	tests := []struct {
		name             string
		idempotencyKey   string
		senderUserID     string
		senderWalletID   string
		receiverUserID   string
		receiverWalletID string
		currency         string
		amount           int64
		wantErr          error
	}{
		{
			name:             "valid request",
			idempotencyKey:   "intent-123",
			senderUserID:     senderUserID,
			senderWalletID:   senderWalletID,
			receiverUserID:   receiverUserID,
			receiverWalletID: receiverWalletID,
			currency:         "IDR",
			amount:           5000,
		},
		{
			name:             "trims surrounding whitespace",
			idempotencyKey:   "  intent-123  ",
			senderUserID:     senderUserID,
			senderWalletID:   senderWalletID,
			receiverUserID:   receiverUserID,
			receiverWalletID: receiverWalletID,
			currency:         "  IDR  ",
			amount:           5000,
		},
		{
			name:             "empty idempotency key",
			idempotencyKey:   "",
			senderUserID:     senderUserID,
			senderWalletID:   senderWalletID,
			receiverUserID:   receiverUserID,
			receiverWalletID: receiverWalletID,
			currency:         "IDR",
			amount:           5000,
			wantErr:          ErrInvalidIdempotencyKey,
		},
		{
			name:             "whitespace-only idempotency key",
			idempotencyKey:   "   ",
			senderUserID:     senderUserID,
			senderWalletID:   senderWalletID,
			receiverUserID:   receiverUserID,
			receiverWalletID: receiverWalletID,
			currency:         "IDR",
			amount:           5000,
			wantErr:          ErrInvalidIdempotencyKey,
		},
		{
			name:             "sender user id not a uuid",
			idempotencyKey:   "intent-123",
			senderUserID:     "not-a-uuid",
			senderWalletID:   senderWalletID,
			receiverUserID:   receiverUserID,
			receiverWalletID: receiverWalletID,
			currency:         "IDR",
			amount:           5000,
			wantErr:          ErrInvalidID,
		},
		{
			name:             "sender wallet id not a uuid",
			idempotencyKey:   "intent-123",
			senderUserID:     senderUserID,
			senderWalletID:   "not-a-uuid",
			receiverUserID:   receiverUserID,
			receiverWalletID: receiverWalletID,
			currency:         "IDR",
			amount:           5000,
			wantErr:          ErrInvalidID,
		},
		{
			name:             "receiver user id not a uuid",
			idempotencyKey:   "intent-123",
			senderUserID:     senderUserID,
			senderWalletID:   senderWalletID,
			receiverUserID:   "not-a-uuid",
			receiverWalletID: receiverWalletID,
			currency:         "IDR",
			amount:           5000,
			wantErr:          ErrInvalidID,
		},
		{
			name:             "receiver wallet id not a uuid",
			idempotencyKey:   "intent-123",
			senderUserID:     senderUserID,
			senderWalletID:   senderWalletID,
			receiverUserID:   receiverUserID,
			receiverWalletID: "not-a-uuid",
			currency:         "IDR",
			amount:           5000,
			wantErr:          ErrInvalidID,
		},
		{
			name:             "sender is receiver",
			idempotencyKey:   "intent-123",
			senderUserID:     senderUserID,
			senderWalletID:   senderWalletID,
			receiverUserID:   senderUserID,
			receiverWalletID: receiverWalletID,
			currency:         "IDR",
			amount:           5000,
			wantErr:          ErrSenderIsReceiver,
		},
		{
			name:             "zero amount",
			idempotencyKey:   "intent-123",
			senderUserID:     senderUserID,
			senderWalletID:   senderWalletID,
			receiverUserID:   receiverUserID,
			receiverWalletID: receiverWalletID,
			currency:         "IDR",
			amount:           0,
			wantErr:          ErrInvalidAmount,
		},
		{
			name:             "negative amount",
			idempotencyKey:   "intent-123",
			senderUserID:     senderUserID,
			senderWalletID:   senderWalletID,
			receiverUserID:   receiverUserID,
			receiverWalletID: receiverWalletID,
			currency:         "IDR",
			amount:           -5000,
			wantErr:          ErrInvalidAmount,
		},
		{
			name:             "empty currency",
			idempotencyKey:   "intent-123",
			senderUserID:     senderUserID,
			senderWalletID:   senderWalletID,
			receiverUserID:   receiverUserID,
			receiverWalletID: receiverWalletID,
			currency:         "",
			amount:           5000,
			wantErr:          ErrInvalidCurrency,
		},
		{
			name:             "whitespace-only currency",
			idempotencyKey:   "intent-123",
			senderUserID:     senderUserID,
			senderWalletID:   senderWalletID,
			receiverUserID:   receiverUserID,
			receiverWalletID: receiverWalletID,
			currency:         "   ",
			amount:           5000,
			wantErr:          ErrInvalidCurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTransactionRequest(tt.idempotencyKey, tt.senderUserID, tt.senderWalletID, tt.receiverUserID, tt.receiverWalletID, tt.currency, tt.amount)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("NewTransactionRequest() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewTransactionRequest() error = %v, want nil", err)
			}
			if got.IdempotencyKey != "intent-123" {
				t.Fatalf("NewTransactionRequest() did not trim idempotency key, got %q", got.IdempotencyKey)
			}
			if got.Currency != "IDR" {
				t.Fatalf("NewTransactionRequest() did not trim currency, got %q", got.Currency)
			}
			if got.Amount != tt.amount {
				t.Fatalf("NewTransactionRequest() amount = %d, want %d", got.Amount, tt.amount)
			}
		})
	}
}

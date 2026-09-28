package eventconsumer

import (
	"context"
	"testing"

	"github.com/segmentio/kafka-go"
)

func TestUserCreatedProcessor_RejectsMalformedPayload(t *testing.T) {
	wallets := &stubWalletCreator{}
	p := NewUserCreatedProcessor(wallets)

	err := p.Process(context.Background(), kafka.Message{Value: []byte("not json at all")})
	if err == nil {
		t.Fatalf("Process() error = nil, want an error for a malformed payload")
	}
	if wallets.callCount != 0 {
		t.Fatalf("CreateWallet was called %d times for a malformed payload, want 0", wallets.callCount)
	}
}

func TestUserCreatedProcessor_RejectsMissingUserID(t *testing.T) {
	wallets := &stubWalletCreator{}
	p := NewUserCreatedProcessor(wallets)

	err := p.Process(context.Background(), kafka.Message{Value: mustMarshalUserCreated(t, "")})
	if err == nil {
		t.Fatalf("Process() error = nil, want an error for an event with no user_id")
	}
	if wallets.callCount != 0 {
		t.Fatalf("CreateWallet was called %d times for an event with no user_id, want 0", wallets.callCount)
	}
}

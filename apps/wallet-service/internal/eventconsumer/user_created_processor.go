package eventconsumer

import (
	"context"
	"errors"
	"fmt"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/encoding/protojson"

	userv1 "github.com/jrmygp/kimo-wallet/apps/wallet-service/gen/user/v1"
	"github.com/jrmygp/kimo-wallet/apps/wallet-service/internal/domain"
)

// WalletCreator is the subset of service.WalletService this processor
// needs.
type WalletCreator interface {
	CreateWallet(ctx context.Context, userID string) (domain.Wallet, error)
}

// UserCreatedProcessor handles "user.created" events: provisioning a
// wallet for the newly registered user. See Handler for the retry/DLQ
// machinery this runs under.
type UserCreatedProcessor struct {
	wallets WalletCreator
}

func NewUserCreatedProcessor(wallets WalletCreator) *UserCreatedProcessor {
	return &UserCreatedProcessor{wallets: wallets}
}

func (p *UserCreatedProcessor) Process(ctx context.Context, msg kafka.Message) error {
	var event userv1.UserCreated
	if err := protojson.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal user.created payload: %w", err)
	}
	if event.GetUserId() == "" {
		return errors.New("user.created event missing user_id")
	}

	if _, err := p.wallets.CreateWallet(ctx, event.GetUserId()); err != nil {
		return fmt.Errorf("create wallet for user %s: %w", event.GetUserId(), err)
	}
	return nil
}

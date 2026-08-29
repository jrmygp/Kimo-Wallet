// Package service orchestrates domain validation, id generation, token
// minting, and persistence.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/domain"
	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/idgen"
)

// UserRepository persists User records.
type UserRepository interface {
	Create(ctx context.Context, id, kimoID string, input domain.RegisterInput) (domain.User, error)
	Login(ctx context.Context, input domain.LoginInput) (domain.User, error)
	// GetUserByID looks a user up by KimoID, not the internal id — see
	// domain.User.KimoID and the RPC/proto doc comments.
	GetUserByID(ctx context.Context, kimoID string) (domain.User, error)
}

// TokenMinter signs a short-lived access token for an authenticated user.
type TokenMinter interface {
	Mint(userID, phoneNumber string) (string, error)
}

// maxKimoIDGenerationAttempts bounds Register's retry loop when a freshly
// generated KimoID collides with an existing one. A KimoID has far less
// keyspace than the UUID `id` (36^12 vs. a full UUID's 122 bits), so unlike
// id generation this is given an explicit, bounded retry rather than
// trusting a single attempt — bounded so a real bug (e.g. a broken
// generator) fails loudly instead of retrying forever, per
// docs/CLAUDE.md §3.6 rule 3's "never an infinite tight retry loop", which
// applies just as much to a generation retry as an event-consumer one.
const maxKimoIDGenerationAttempts = 5

type UserService struct {
	repo   UserRepository
	minter TokenMinter
}

func NewUserService(repo UserRepository, minter TokenMinter) *UserService {
	return &UserService{repo: repo, minter: minter}
}

// Register validates the input, mints a new user id and KimoID, persists
// the user, and signs an access token for the newly created user. Returns
// domain.ErrInvalidPhoneNumber / domain.ErrInvalidFullName on validation
// failure, or domain.ErrPhoneNumberTaken if the phone number is already
// registered.
func (s *UserService) Register(ctx context.Context, phoneNumber, fullName string) (domain.User, string, error) {
	input, err := domain.NewRegisterInput(phoneNumber, fullName)
	if err != nil {
		return domain.User{}, "", err
	}

	id, err := idgen.NewV4()
	if err != nil {
		return domain.User{}, "", fmt.Errorf("register user: %w", err)
	}

	var user domain.User
	for attempt := 1; ; attempt++ {
		kimoID, err := idgen.NewKimoID()
		if err != nil {
			return domain.User{}, "", fmt.Errorf("register user: %w", err)
		}

		user, err = s.repo.Create(ctx, id, kimoID, input)
		if err == nil {
			break
		}
		if errors.Is(err, domain.ErrKimoIDCollision) && attempt < maxKimoIDGenerationAttempts {
			continue
		}
		return domain.User{}, "", err
	}

	token, err := s.minter.Mint(user.ID, user.PhoneNumber)
	if err != nil {
		return domain.User{}, "", fmt.Errorf("register user: %w", err)
	}

	return user, token, nil
}

// Login finds the user by phone number and signs an access token for them.
// Returns domain.ErrInvalidPhoneNumber on validation failure, or
// domain.ErrUserNotFound if no user has that phone number.
func (s *UserService) Login(ctx context.Context, phoneNumber string) (domain.User, string, error) {
	input, err := domain.NewLoginInput(phoneNumber)
	if err != nil {
		return domain.User{}, "", err
	}

	user, err := s.repo.Login(ctx, input)
	if err != nil {
		return domain.User{}, "", err
	}

	token, err := s.minter.Mint(user.ID, user.PhoneNumber)
	if err != nil {
		return domain.User{}, "", fmt.Errorf("login: %w", err)
	}

	return user, token, nil
}

// GetUserByID looks up a user by their KimoID (not the internal id —
// see domain.User.KimoID). Returns domain.ErrInvalidKimoID if kimoID isn't
// shaped like a KimoID (rejected before it ever reaches the database —
// a generic no-match isn't something callers should have to distinguish
// from a real infrastructure failure), or domain.ErrUserNotFound if it's
// well-formed but no such user exists.
func (s *UserService) GetUserByID(ctx context.Context, kimoID string) (domain.User, error) {
	kimoID, err := domain.NewGetUserByKimoIDInput(kimoID)
	if err != nil {
		return domain.User{}, err
	}

	user, err := s.repo.GetUserByID(ctx, kimoID)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

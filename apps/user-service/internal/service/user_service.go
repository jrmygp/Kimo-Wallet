// Package service orchestrates domain validation, id generation, token
// minting, and persistence.
package service

import (
	"context"
	"fmt"

	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/domain"
	"github.com/jrmygp/kimo-wallet/apps/user-service/internal/idgen"
)

// UserRepository persists User records.
type UserRepository interface {
	Create(ctx context.Context, id string, input domain.RegisterInput) (domain.User, error)
	Login(ctx context.Context, input domain.LoginInput) (domain.User, error)
	GetUserByID(ctx context.Context, id string) (domain.User, error)
}

// TokenMinter signs a short-lived access token for an authenticated user.
type TokenMinter interface {
	Mint(userID, phoneNumber string) (string, error)
}

type UserService struct {
	repo   UserRepository
	minter TokenMinter
}

func NewUserService(repo UserRepository, minter TokenMinter) *UserService {
	return &UserService{repo: repo, minter: minter}
}

// Register validates the input, mints a new user id, persists the user, and
// signs an access token for the newly created user. Returns
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

	user, err := s.repo.Create(ctx, id, input)
	if err != nil {
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

func (s *UserService) GetUserByID(ctx context.Context, id string) (domain.User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

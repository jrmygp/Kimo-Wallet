// Package service orchestrates domain validation, id generation, and persistence.
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
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Register validates the input, mints a new user id, and persists the user.
// Returns domain.ErrInvalidPhoneNumber / domain.ErrInvalidFullName on validation
// failure, or domain.ErrPhoneNumberTaken if the phone number is already registered.
func (s *UserService) Register(ctx context.Context, phoneNumber, fullName string) (domain.User, error) {
	input, err := domain.NewRegisterInput(phoneNumber, fullName)
	if err != nil {
		return domain.User{}, err
	}

	id, err := idgen.NewV4()
	if err != nil {
		return domain.User{}, fmt.Errorf("register user: %w", err)
	}

	user, err := s.repo.Create(ctx, id, input)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

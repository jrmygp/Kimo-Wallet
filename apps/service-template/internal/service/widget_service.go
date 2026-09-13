// Package service orchestrates id generation, input validation, and
// persistence for widgets — the layer grpcserver calls into and
// storage/postgres implements against.
package service

import (
	"context"
	"fmt"

	"github.com/jrmygp/kimo-wallet/apps/service-template/internal/domain"
	"github.com/jrmygp/kimo-wallet/apps/service-template/internal/idgen"
)

// WidgetRepository persists Widget records.
type WidgetRepository interface {
	Create(ctx context.Context, id string, input domain.CreateWidgetInput) (domain.Widget, error)
	GetByID(ctx context.Context, id string) (domain.Widget, error)
}

type WidgetService struct {
	repo WidgetRepository
}

func NewWidgetService(repo WidgetRepository) *WidgetService {
	return &WidgetService{repo: repo}
}

func (s *WidgetService) CreateWidget(ctx context.Context, name string) (domain.Widget, error) {
	input, err := domain.NewCreateWidgetInput(name)
	if err != nil {
		return domain.Widget{}, err
	}

	id, err := idgen.NewV4()
	if err != nil {
		return domain.Widget{}, fmt.Errorf("create widget: %w", err)
	}

	return s.repo.Create(ctx, id, input)
}

func (s *WidgetService) GetWidgetByID(ctx context.Context, id string) (domain.Widget, error) {
	validID, err := domain.NewGetWidgetByIDInput(id)
	if err != nil {
		return domain.Widget{}, err
	}

	return s.repo.GetByID(ctx, validID)
}

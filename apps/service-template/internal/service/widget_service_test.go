package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jrmygp/kimo-wallet/apps/service-template/internal/domain"
)

type stubWidgetRepository struct {
	createErr     error
	createdWidget domain.Widget
	getByIDErr    error
	gotWidget     domain.Widget
	// createCalls records every name Create was called with.
	createCalls []string
}

func (s *stubWidgetRepository) Create(ctx context.Context, id string, input domain.CreateWidgetInput) (domain.Widget, error) {
	s.createCalls = append(s.createCalls, input.Name)
	if s.createErr != nil {
		return domain.Widget{}, s.createErr
	}
	return s.createdWidget, nil
}

func (s *stubWidgetRepository) GetByID(ctx context.Context, id string) (domain.Widget, error) {
	if s.getByIDErr != nil {
		return domain.Widget{}, s.getByIDErr
	}
	return s.gotWidget, nil
}

func TestWidgetService_CreateWidget_Success(t *testing.T) {
	want := domain.Widget{ID: "widget-1", Name: "My Widget", Status: "ACTIVE"}
	repo := &stubWidgetRepository{createdWidget: want}
	svc := NewWidgetService(repo)

	got, err := svc.CreateWidget(context.Background(), "My Widget")
	if err != nil {
		t.Fatalf("CreateWidget() error = %v, want nil", err)
	}
	if got != want {
		t.Fatalf("CreateWidget() = %+v, want %+v", got, want)
	}
}

func TestWidgetService_CreateWidget_InvalidName_NeverReachesRepository(t *testing.T) {
	repo := &stubWidgetRepository{}
	svc := NewWidgetService(repo)

	_, err := svc.CreateWidget(context.Background(), "   ")
	if !errors.Is(err, domain.ErrInvalidName) {
		t.Fatalf("CreateWidget() error = %v, want %v", err, domain.ErrInvalidName)
	}
	if len(repo.createCalls) != 0 {
		t.Fatalf("repository Create() was called %d times for an invalid name, want 0", len(repo.createCalls))
	}
}

func TestWidgetService_GetWidgetByID_InvalidID_NeverReachesRepository(t *testing.T) {
	repo := &stubWidgetRepository{}
	svc := NewWidgetService(repo)

	_, err := svc.GetWidgetByID(context.Background(), "not-a-uuid")
	if !errors.Is(err, domain.ErrInvalidID) {
		t.Fatalf("GetWidgetByID() error = %v, want %v", err, domain.ErrInvalidID)
	}
}

func TestWidgetService_GetWidgetByID_NotFound(t *testing.T) {
	repo := &stubWidgetRepository{getByIDErr: domain.ErrWidgetNotFound}
	svc := NewWidgetService(repo)

	_, err := svc.GetWidgetByID(context.Background(), "11111111-1111-4111-8111-111111111111")
	if !errors.Is(err, domain.ErrWidgetNotFound) {
		t.Fatalf("GetWidgetByID() error = %v, want %v", err, domain.ErrWidgetNotFound)
	}
}

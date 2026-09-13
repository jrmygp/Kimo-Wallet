package grpcserver

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	widgetv1 "github.com/jrmygp/kimo-wallet/apps/service-template/gen/widget/v1"
	"github.com/jrmygp/kimo-wallet/apps/service-template/internal/domain"
)

// widgetService is the subset of *service.WidgetService this server
// depends on, so it can be exercised in tests with a stub instead of a
// real service instance.
type widgetService interface {
	CreateWidget(ctx context.Context, name string) (domain.Widget, error)
	GetWidgetByID(ctx context.Context, id string) (domain.Widget, error)
}

func toProtoWidget(widget domain.Widget) *widgetv1.Widget {
	return &widgetv1.Widget{
		Id:        widget.ID,
		Name:      widget.Name,
		Status:    widget.Status,
		CreatedAt: timestamppb.New(widget.CreatedAt),
	}
}

type WidgetServer struct {
	widgetv1.UnimplementedWidgetServiceServer
	service widgetService
}

func NewWidgetServer(service widgetService) *WidgetServer {
	return &WidgetServer{service: service}
}

func (s *WidgetServer) CreateWidget(ctx context.Context, req *widgetv1.CreateWidgetRequest) (*widgetv1.CreateWidgetResponse, error) {
	widget, err := s.service.CreateWidget(ctx, req.GetName())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidName):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to create widget")
		}
	}

	return &widgetv1.CreateWidgetResponse{Widget: toProtoWidget(widget)}, nil
}

func (s *WidgetServer) GetWidgetByID(ctx context.Context, req *widgetv1.GetWidgetByIDRequest) (*widgetv1.GetWidgetByIDResponse, error) {
	widget, err := s.service.GetWidgetByID(ctx, req.GetId())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidID):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, domain.ErrWidgetNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to find widget")
		}
	}

	return &widgetv1.GetWidgetByIDResponse{Widget: toProtoWidget(widget)}, nil
}

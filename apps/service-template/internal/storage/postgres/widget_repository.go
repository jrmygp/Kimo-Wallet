package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/jrmygp/kimo-wallet/apps/service-template/internal/domain"
)

const statusActive = "ACTIVE"

// widgetModel is the GORM row mapping for the widgets table. Kept
// separate from domain.Widget so the domain package stays free of ORM
// struct tags — same convention as every other service in this repo
// (docs/CLAUDE.md §5.4).
type widgetModel struct {
	ID        string `gorm:"primaryKey"`
	Name      string
	Status    string
	CreatedAt time.Time // populated by GORM on Create via its CreatedAt convention
}

func (widgetModel) TableName() string { return "widgets" }

type WidgetRepository struct {
	db *gorm.DB
}

func NewWidgetRepository(db *gorm.DB) *WidgetRepository {
	return &WidgetRepository{db: db}
}

func (r *WidgetRepository) Create(ctx context.Context, id string, input domain.CreateWidgetInput) (domain.Widget, error) {
	row := widgetModel{
		ID:     id,
		Name:   input.Name,
		Status: statusActive,
	}

	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Widget{}, fmt.Errorf("insert widget: %w", err)
	}

	return domain.Widget{
		ID:        row.ID,
		Name:      row.Name,
		Status:    row.Status,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *WidgetRepository) GetByID(ctx context.Context, id string) (domain.Widget, error) {
	var row widgetModel

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Widget{}, domain.ErrWidgetNotFound
		}
		return domain.Widget{}, fmt.Errorf("find widget by id: %w", err)
	}

	return domain.Widget{
		ID:        row.ID,
		Name:      row.Name,
		Status:    row.Status,
		CreatedAt: row.CreatedAt,
	}, nil
}

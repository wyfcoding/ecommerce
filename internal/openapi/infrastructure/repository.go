package infrastructure

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/openapi/domain"
	"gorm.io/gorm"
)

type GormOpenApiRepository struct {
	db *gorm.DB
}

func NewGormOpenApiRepository(db *gorm.DB) *GormOpenApiRepository {
	return &GormOpenApiRepository{db: db}
}

func (r *GormOpenApiRepository) SaveApp(ctx context.Context, app *domain.OpenApiApp) error {
	return r.db.WithContext(ctx).Save(app).Error
}

func (r *GormOpenApiRepository) GetAppByID(ctx context.Context, appID string) (*domain.OpenApiApp, error) {
	var app domain.OpenApiApp
	err := r.db.WithContext(ctx).Where("app_id = ?", appID).First(&app).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &app, err
}

func (r *GormOpenApiRepository) GetAppByKey(ctx context.Context, apiKey string) (*domain.OpenApiApp, error) {
	var app domain.OpenApiApp
	err := r.db.WithContext(ctx).Where("api_key = ?", apiKey).First(&app).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &app, err
}

func (r *GormOpenApiRepository) ListAppsByUserID(ctx context.Context, userID string) ([]*domain.OpenApiApp, error) {
	var apps []*domain.OpenApiApp
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&apps).Error
	return apps, err
}

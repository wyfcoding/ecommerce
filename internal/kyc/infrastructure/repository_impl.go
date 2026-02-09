package infrastructure

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/kyc/domain"
	"gorm.io/gorm"
)

type KYCRepositoryImpl struct {
	db *gorm.DB
}

func NewKYCRepository(db *gorm.DB) domain.KYCRepository {
	return &KYCRepositoryImpl{db: db}
}

func (r *KYCRepositoryImpl) Save(ctx context.Context, app *domain.KYCApplication) error {
	return r.db.WithContext(ctx).Save(app).Error
}

func (r *KYCRepositoryImpl) FindByUserID(ctx context.Context, userID uint64) (*domain.KYCApplication, error) {
	var app domain.KYCApplication
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *KYCRepositoryImpl) FindByID(ctx context.Context, appID string) (*domain.KYCApplication, error) {
	var app domain.KYCApplication
	if err := r.db.WithContext(ctx).Where("application_id = ?", appID).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

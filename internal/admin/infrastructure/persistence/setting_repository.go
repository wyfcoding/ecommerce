package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
	"gorm.io/gorm"
)

type settingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) domain.SettingRepository {
	return &settingRepository{db: db}
}

func (r *settingRepository) GetByKey(ctx context.Context, key string) (*domain.SystemSetting, error) {
	var setting domain.SystemSetting
	if err := r.db.WithContext(ctx).Where("`key` = ?", key).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *settingRepository) Save(ctx context.Context, setting *domain.SystemSetting) error {
	return r.db.WithContext(ctx).Save(setting).Error
}

package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"ecommerce/internal/aftersales/model"
)

// AftersalesRepository 定义了售后数据仓库的接口
type AftersalesRepository interface {
	CreateApplication(ctx context.Context, app *model.AftersalesApplication) error
	GetApplicationByID(ctx context.Context, id uint) (*model.AftersalesApplication, error)
	GetApplicationBySN(ctx context.Context, sn string) (*model.AftersalesApplication, error)
	ListApplicationsByUserID(ctx context.Context, userID uint) ([]model.AftersalesApplication, error)
	UpdateApplication(ctx context.Context, app *model.AftersalesApplication) error
}

// aftersalesRepository 是接口的具体实现
type aftersalesRepository struct {
	db *gorm.DB
}

// NewAftersalesRepository 创建一个新的 aftersalesRepository 实例
func NewAftersalesRepository(db *gorm.DB) AftersalesRepository {
	return &aftersalesRepository{db: db}
}

// CreateApplication 在事务中创建售后申请及其关联的项目
func (r *aftersalesRepository) CreateApplication(ctx context.Context, app *model.AftersalesApplication) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(app).Error; err != nil {
			return fmt.Errorf("数据库创建售后申请失败: %w", err)
		}
		// GORM 的关联创建会自动处理 app.Items
		return nil
	})
}

// GetApplicationByID 根据 ID 获取申请详情
func (r *aftersalesRepository) GetApplicationByID(ctx context.Context, id uint) (*model.AftersalesApplication, error) {
	var app model.AftersalesApplication
	if err := r.db.WithContext(ctx).Preload("Items").First(&app, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("数据库查询售后申请失败: %w", err)
	}
	return &app, nil
}

// GetApplicationBySN 根据 SN 获取申请详情
func (r *aftersalesRepository) GetApplicationBySN(ctx context.Context, sn string) (*model.AftersalesApplication, error) {
	var app model.AftersalesApplication
	if err := r.db.WithContext(ctx).Preload("Items").Where("application_sn = ?", sn).First(&app).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("数据库查询售后申请失败: %w", err)
	}
	return &app, nil
}

// ListApplicationsByUserID 列出某个用户的所有售后申请
func (r *aftersalesRepository) ListApplicationsByUserID(ctx context.Context, userID uint) ([]model.AftersalesApplication, error) {
	var apps []model.AftersalesApplication
	if err := r.db.WithContext(ctx).Preload("Items").Where("user_id = ?", userID).Order("created_at desc").Find(&apps).Error; err != nil {
		return nil, fmt.Errorf("数据库列出售后申请失败: %w", err)
	}
	return apps, nil
}

// UpdateApplication 更新售后申请信息 (例如，状态、备注等)
func (r *aftersalesRepository) UpdateApplication(ctx context.Context, app *model.AftersalesApplication) error {
	if err := r.db.WithContext(ctx).Save(app).Error; err != nil {
		return fmt.Errorf("数据库更新售后申请失败: %w", err)
	}
	return nil
}

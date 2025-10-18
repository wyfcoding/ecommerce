package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"ecommerce/internal/asset/model"
)

// AssetRepository 定义了资产数据仓库的接口
type AssetRepository interface {
	CreateAsset(ctx context.Context, asset *model.Asset) error
	GetAssetByID(ctx context.Context, id uint) (*model.Asset, error)
	GetAssetByObjectKey(ctx context.Context, objectKey string) (*model.Asset, error)
	ListAssetsByRelatedID(ctx context.Context, assetType model.AssetType, relatedID uint) ([]model.Asset, error)
	DeleteAsset(ctx context.Context, id uint) error // 软删除
}

// assetRepository 是接口的具体实现
type assetRepository struct {
	db *gorm.DB
}

// NewAssetRepository 创建一个新的 assetRepository 实例
func NewAssetRepository(db *gorm.DB) AssetRepository {
	return &assetRepository{db: db}
}

// CreateAsset 在数据库中创建一条新的资产元数据记录
func (r *assetRepository) CreateAsset(ctx context.Context, asset *model.Asset) error {
	if err := r.db.WithContext(ctx).Create(asset).Error; err != nil {
		return fmt.Errorf("数据库创建资产元数据失败: %w", err)
	}
	return nil
}

// GetAssetByID 根据 ID 获取资产元数据
func (r *assetRepository) GetAssetByID(ctx context.Context, id uint) (*model.Asset, error) {
	var asset model.Asset
	if err := r.db.WithContext(ctx).First(&asset, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("数据库查询资产失败: %w", err)
	}
	return &asset, nil
}

// GetAssetByObjectKey 根据对象存储中的 Key 获取资产元数据
func (r *assetRepository) GetAssetByObjectKey(ctx context.Context, objectKey string) (*model.Asset, error) {
	var asset model.Asset
	if err := r.db.WithContext(ctx).Where("object_key = ?", objectKey).First(&asset).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("数据库查询资产失败: %w", err)
	}
	return &asset, nil
}

// ListAssetsByRelatedID 列出与特定实体关联的所有资产
func (r *assetRepository) ListAssetsByRelatedID(ctx context.Context, assetType model.AssetType, relatedID uint) ([]model.Asset, error) {
	var assets []model.Asset
	if err := r.db.WithContext(ctx).Where("asset_type = ? AND related_id = ?", assetType, relatedID).Find(&assets).Error; err != nil {
		return nil, fmt.Errorf("数据库列出关联资产失败: %w", err)
	}
	return assets, nil
}

// DeleteAsset 软删除资产元数据
func (r *assetRepository) DeleteAsset(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.Asset{}, id).Error; err != nil {
		return fmt.Errorf("数据库删除资产失败: %w", err)
	}
	return nil
}
package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/product/domain"

	"gorm.io/gorm"
)

type SKURepository struct {
	db *gorm.DB
}

func NewSKURepository(db *gorm.DB) *SKURepository {
	return &SKURepository{db: db}
}

func (r *SKURepository) Save(ctx context.Context, sku *domain.SKU) error {
	return r.db.WithContext(ctx).Create(sku).Error
}

func (r *SKURepository) FindByID(ctx context.Context, id uint) (*domain.SKU, error) {
	var sku domain.SKU
	if err := r.db.WithContext(ctx).First(&sku, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sku, nil
}

func (r *SKURepository) FindByProductID(ctx context.Context, productID uint) ([]*domain.SKU, error) {
	var skus []*domain.SKU
	if err := r.db.WithContext(ctx).Where("product_id = ?", productID).Find(&skus).Error; err != nil {
		return nil, err
	}
	return skus, nil
}

func (r *SKURepository) Update(ctx context.Context, sku *domain.SKU) error {
	return r.db.WithContext(ctx).Save(sku).Error
}

func (r *SKURepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.SKU{}, id).Error
}

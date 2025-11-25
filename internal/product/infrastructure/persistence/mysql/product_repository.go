package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/product/domain"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Save(ctx context.Context, product *domain.Product) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(product).Error; err != nil {
			return err
		}
		for _, sku := range product.SKUs {
			sku.ProductID = product.ID
			if err := tx.Create(sku).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ProductRepository) FindByID(ctx context.Context, id uint) (*domain.Product, error) {
	var product domain.Product
	if err := r.db.WithContext(ctx).Preload("SKUs").First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) FindByName(ctx context.Context, name string) (*domain.Product, error) {
	var product domain.Product
	if err := r.db.WithContext(ctx).Preload("SKUs").Where("name = ?", name).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(product).Error; err != nil {
			return err
		}
		// Update SKUs logic if needed, usually SKUs are updated separately or replaced
		return nil
	})
}

func (r *ProductRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Product{}, id).Error
}

func (r *ProductRepository) List(ctx context.Context, offset, limit int) ([]*domain.Product, int64, error) {
	var products []*domain.Product
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Preload("SKUs").Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductRepository) ListByCategory(ctx context.Context, categoryID uint, offset, limit int) ([]*domain.Product, int64, error) {
	var products []*domain.Product
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.Product{}).Where("category_id = ?", categoryID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Preload("SKUs").Where("category_id = ?", categoryID).Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductRepository) ListByBrand(ctx context.Context, brandID uint, offset, limit int) ([]*domain.Product, int64, error) {
	var products []*domain.Product
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.Product{}).Where("brand_id = ?", brandID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Preload("SKUs").Where("brand_id = ?", brandID).Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

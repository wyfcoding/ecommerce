package mysql

import (
	"context"
	"errors"

	"ecommerce/internal/product/domain"

	"gorm.io/gorm"
)

type BrandRepository struct {
	db *gorm.DB
}

func NewBrandRepository(db *gorm.DB) *BrandRepository {
	return &BrandRepository{db: db}
}

func (r *BrandRepository) Save(ctx context.Context, brand *domain.Brand) error {
	return r.db.WithContext(ctx).Create(brand).Error
}

func (r *BrandRepository) FindByID(ctx context.Context, id uint) (*domain.Brand, error) {
	var brand domain.Brand
	if err := r.db.WithContext(ctx).First(&brand, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &brand, nil
}

func (r *BrandRepository) FindByName(ctx context.Context, name string) (*domain.Brand, error) {
	var brand domain.Brand
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&brand).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &brand, nil
}

func (r *BrandRepository) Update(ctx context.Context, brand *domain.Brand) error {
	return r.db.WithContext(ctx).Save(brand).Error
}

func (r *BrandRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Brand{}, id).Error
}

func (r *BrandRepository) List(ctx context.Context) ([]*domain.Brand, error) {
	var brands []*domain.Brand
	if err := r.db.WithContext(ctx).Find(&brands).Error; err != nil {
		return nil, err
	}
	return brands, nil
}

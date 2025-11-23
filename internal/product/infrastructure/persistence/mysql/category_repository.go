package mysql

import (
	"context"
	"errors"

	"ecommerce/internal/product/domain"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Save(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *CategoryRepository) FindByID(ctx context.Context, id uint) (*domain.Category, error) {
	var category domain.Category
	if err := r.db.WithContext(ctx).First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) FindByName(ctx context.Context, name string) (*domain.Category, error) {
	var category domain.Category
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *CategoryRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Category{}, id).Error
}

func (r *CategoryRepository) List(ctx context.Context) ([]*domain.Category, error) {
	var categories []*domain.Category
	if err := r.db.WithContext(ctx).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *CategoryRepository) FindByParentID(ctx context.Context, parentID uint) ([]*domain.Category, error) {
	var categories []*domain.Category
	if err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

package repository

import (
	"context"

	"ecommerce/internal/product/model"
)

// BrandRepo is a Brand repo.
type BrandRepo interface {
	CreateBrand(ctx context.Context, brand *model.Brand) (*model.Brand, error)
	UpdateBrand(ctx context.Context, brand *model.Brand) (*model.Brand, error)
	DeleteBrand(ctx context.Context, id uint64) error
	ListBrands(ctx context.Context, pageSize, pageNum uint32, name *string, isVisible *bool) ([]*model.Brand, uint64, error)
}

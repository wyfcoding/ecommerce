package repository

import (
	"context"

	"ecommerce/internal/product/model"
)

// CategoryRepo 定义了分类数据仓库的接口。
type CategoryRepo interface {
	CreateCategory(ctx context.Context, c *model.Category) (*model.Category, error)
	UpdateCategory(ctx context.Context, c *model.Category) (*model.Category, error)
	DeleteCategory(ctx context.Context, id uint64) error
	GetCategory(ctx context.Context, id uint64) (*model.Category, error)
	ListCategories(ctx context.Context, parentID uint64) ([]*model.Category, error)
}
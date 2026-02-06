package domain

import "context"

// ProductReadRepository 定义商品读模型的高性能访问接口（Redis）。
type ProductReadRepository interface {
	Save(ctx context.Context, product *Product) error
	GetByID(ctx context.Context, id uint64) (*Product, error)
	Delete(ctx context.Context, id uint64) error
}

// SKUReadRepository 定义 SKU 读模型接口（Redis）。
type SKUReadRepository interface {
	Save(ctx context.Context, sku *SKU) error
	GetByID(ctx context.Context, id uint64) (*SKU, error)
	Delete(ctx context.Context, id uint64) error
}

// BrandReadRepository 定义品牌读模型接口（Redis）。
type BrandReadRepository interface {
	Save(ctx context.Context, brand *Brand) error
	GetByID(ctx context.Context, id uint64) (*Brand, error)
	Delete(ctx context.Context, id uint64) error
}

// CategoryReadRepository 定义分类读模型接口（Redis）。
type CategoryReadRepository interface {
	Save(ctx context.Context, category *Category) error
	GetByID(ctx context.Context, id uint64) (*Category, error)
	Delete(ctx context.Context, id uint64) error
}

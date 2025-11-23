package domain

import "context"

// ProductRepository 商品仓储接口
type ProductRepository interface {
	Save(ctx context.Context, product *Product) error
	FindByID(ctx context.Context, id uint) (*Product, error)
	FindByName(ctx context.Context, name string) (*Product, error)
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int) ([]*Product, int64, error)
	ListByCategory(ctx context.Context, categoryID uint, offset, limit int) ([]*Product, int64, error)
	ListByBrand(ctx context.Context, brandID uint, offset, limit int) ([]*Product, int64, error)
}

// SKURepository SKU仓储接口
type SKURepository interface {
	Save(ctx context.Context, sku *SKU) error
	FindByID(ctx context.Context, id uint) (*SKU, error)
	FindByProductID(ctx context.Context, productID uint) ([]*SKU, error)
	Update(ctx context.Context, sku *SKU) error
	Delete(ctx context.Context, id uint) error
}

// CategoryRepository 分类仓储接口
type CategoryRepository interface {
	Save(ctx context.Context, category *Category) error
	FindByID(ctx context.Context, id uint) (*Category, error)
	FindByName(ctx context.Context, name string) (*Category, error)
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]*Category, error)
	FindByParentID(ctx context.Context, parentID uint) ([]*Category, error)
}

// BrandRepository 品牌仓储接口
type BrandRepository interface {
	Save(ctx context.Context, brand *Brand) error
	FindByID(ctx context.Context, id uint) (*Brand, error)
	FindByName(ctx context.Context, name string) (*Brand, error)
	Update(ctx context.Context, brand *Brand) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]*Brand, error)
}

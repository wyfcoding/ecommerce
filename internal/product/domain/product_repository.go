package domain

import "context"

// ProductRepository 是商品模块的仓储接口。
// 它定义了对 Product 实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type ProductRepository interface {
	// Save 将商品实体保存到数据存储中。
	// 如果商品已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// product: 待保存的商品实体。
	Save(ctx context.Context, product *Product) error
	// FindByID 根据ID获取商品实体。
	FindByID(ctx context.Context, id uint) (*Product, error)
	// FindByName 根据名称获取商品实体。
	FindByName(ctx context.Context, name string) (*Product, error)
	// Update 更新商品实体。
	Update(ctx context.Context, product *Product) error
	// Delete 根据ID删除商品实体。
	Delete(ctx context.Context, id uint) error
	// List 列出所有商品实体，支持分页。
	List(ctx context.Context, offset, limit int) ([]*Product, int64, error)
	// ListByCategory 列出指定分类ID下的商品实体，支持分页。
	ListByCategory(ctx context.Context, categoryID uint, offset, limit int) ([]*Product, int64, error)
	// ListByBrand 列出指定品牌ID下的商品实体，支持分页。
	ListByBrand(ctx context.Context, brandID uint, offset, limit int) ([]*Product, int64, error)
}

// SKURepository 是商品SKU模块的仓储接口。
// 它定义了对 SKU 实体进行数据持久化操作的契约。
type SKURepository interface {
	// Save 将SKU实体保存到数据存储中。
	Save(ctx context.Context, sku *SKU) error
	// FindByID 根据ID获取SKU实体。
	FindByID(ctx context.Context, id uint) (*SKU, error)
	// FindByProductID 根据商品ID获取所有关联的SKU实体。
	FindByProductID(ctx context.Context, productID uint) ([]*SKU, error)
	// Update 更新SKU实体。
	Update(ctx context.Context, sku *SKU) error
	// Delete 根据ID删除SKU实体。
	Delete(ctx context.Context, id uint) error
}

// CategoryRepository 是商品分类模块的仓储接口。
// 它定义了对 Category 实体进行数据持久化操作的契约。
type CategoryRepository interface {
	// Save 将分类实体保存到数据存储中。
	Save(ctx context.Context, category *Category) error
	// FindByID 根据ID获取分类实体。
	FindByID(ctx context.Context, id uint) (*Category, error)
	// FindByName 根据名称获取分类实体。
	FindByName(ctx context.Context, name string) (*Category, error)
	// Update 更新分类实体。
	Update(ctx context.Context, category *Category) error
	// Delete 根据ID删除分类实体。
	Delete(ctx context.Context, id uint) error
	// List 列出所有分类实体。
	List(ctx context.Context) ([]*Category, error)
	// FindByParentID 根据父分类ID获取子分类实体列表。
	FindByParentID(ctx context.Context, parentID uint) ([]*Category, error)
}

// BrandRepository 是商品品牌模块的仓储接口。
// 它定义了对 Brand 实体进行数据持久化操作的契约。
type BrandRepository interface {
	// Save 将品牌实体保存到数据存储中。
	Save(ctx context.Context, brand *Brand) error
	// FindByID 根据ID获取品牌实体。
	FindByID(ctx context.Context, id uint) (*Brand, error)
	// FindByName 根据名称获取品牌实体。
	FindByName(ctx context.Context, name string) (*Brand, error)
	// Update 更新品牌实体。
	Update(ctx context.Context, brand *Brand) error
	// Delete 根据ID删除品牌实体。
	Delete(ctx context.Context, id uint) error
	// List 列出所有品牌实体。
	List(ctx context.Context) ([]*Brand, error)
}

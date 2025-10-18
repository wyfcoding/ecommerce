package repository

import (
	"context"
	"fmt"

	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"

	"ecommerce/internal/product/model"
)

// ProductRepository 定义了商品数据仓库的接口
type ProductRepository interface {
	// 数据库操作
	CreateProduct(ctx context.Context, product *model.Product) error
	GetProductByID(ctx context.Context, id uint) (*model.Product, error)
	UpdateProduct(ctx context.Context, product *model.Product) error
	DeleteProduct(ctx context.Context, id uint) error
	ListProducts(ctx context.Context, page, pageSize int, categoryID, brandID *uint) ([]model.Product, int64, error)
	UpdateStock(ctx context.Context, productID uint, quantityChange int) error

	// Elasticsearch 操作 (用于搜索)
	IndexProduct(ctx context.Context, product *model.Product) error
	SearchProducts(ctx context.Context, query string, page, pageSize int) ([]model.Product, int64, error)
	DeleteProductFromIndex(ctx context.Context, productID uint) error
}

// productRepository 是接口的具体实现
type productRepository struct {
	db *gorm.DB
	es *elastic.Client // Elasticsearch 客户端
}

// NewProductRepository 创建一个新的 productRepository 实例
func NewProductRepository(db *gorm.DB, es *elastic.Client) ProductRepository {
	return &productRepository{db: db, es: es}
}

// --- 数据库操作实现 ---

// CreateProduct 在数据库中创建一个新商品
func (r *productRepository) CreateProduct(ctx context.Context, product *model.Product) error {
	if err := r.db.WithContext(ctx).Create(product).Error; err != nil {
		return fmt.Errorf("数据库创建商品失败: %w", err)
	}
	return nil
}

// GetProductByID 从数据库中按 ID 获取商品
func (r *productRepository) GetProductByID(ctx context.Context, id uint) (*model.Product, error) {
	var product model.Product
	// Preload 加载关联的 Category 和 Brand 信息
	if err := r.db.WithContext(ctx).Preload("Category").Preload("Brand").First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 商品不存在是正常情况
		}
		return nil, fmt.Errorf("数据库查询商品失败: %w", err)
	}
	return &product, nil
}

// UpdateProduct 更新数据库中的商品信息
func (r *productRepository) UpdateProduct(ctx context.Context, product *model.Product) error {
	if err := r.db.WithContext(ctx).Save(product).Error; err != nil {
		return fmt.Errorf("数据库更新商品失败: %w", err)
	}
	return nil
}

// DeleteProduct 从数据库中删除一个商品 (软删除)
func (r *productRepository) DeleteProduct(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.Product{}, id).Error; err != nil {
		return fmt.Errorf("数据库删除商品失败: %w", err)
	}
	return nil
}

// ListProducts 分页并按条件列出商品
func (r *productRepository) ListProducts(ctx context.Context, page, pageSize int, categoryID, brandID *uint) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Product{})

	// 应用过滤条件
	if categoryID != nil {
		db = db.Where("category_id = ?", *categoryID)
	}
	if brandID != nil {
		db = db.Where("brand_id = ?", *brandID)
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("数据库统计商品数量失败: %w", err)
	}

	// 计算偏移量并查询
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Preload("Category").Preload("Brand").Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("数据库列出商品失败: %w", err)
	}

	return products, total, nil
}

// UpdateStock 原子地更新商品库存
// quantityChange 可以是正数（入库）或负数（出库）
func (r *productRepository) UpdateStock(ctx context.Context, productID uint, quantityChange int) error {
    // 使用 gorm.Expr 实现原子更新: stock = stock + ?
    // 增加 WHERE 条件防止库存变为负数
    result := r.db.WithContext(ctx).Model(&model.Product{}).
        Where("id = ? AND stock + ? >= 0", productID, quantityChange).
        Update("stock", gorm.Expr("stock + ?", quantityChange))

    if result.Error != nil {
        return fmt.Errorf("数据库更新库存失败: %w", result.Error)
    }
    if result.RowsAffected == 0 {
        return fmt.Errorf("库存不足或商品不存在")
    }
    return nil
}


// --- Elasticsearch 操作实现 ---

// IndexProduct 将商品数据索引到 Elasticsearch
func (r *productRepository) IndexProduct(ctx context.Context, product *model.Product) error {
	// 这里的 "products" 是 Elasticsearch 中的索引名称
	_, err := r.es.Index().
		Index("products").
		Id(fmt.Sprintf("%d", product.ID)).
		BodyJson(product).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("索引商品到 Elasticsearch 失败: %w", err)
	}
	return nil
}

// SearchProducts 在 Elasticsearch 中搜索商品
func (r *productRepository) SearchProducts(ctx context.Context, query string, page, pageSize int) ([]model.Product, int64, error) {
	// ... Elasticsearch 搜索逻辑
	// 这是一个复杂的实现，通常会构建一个 multi_match 查询来搜索多个字段
	// 这里仅为示例，实际代码会更复杂
	// esQuery := elastic.NewMultiMatchQuery(query, "name", "description", "category.name", "brand.name").Fuzziness("AUTO")
	// searchResult, err := r.es.Search()... 
	return nil, 0, fmt.Errorf("搜索功能待实现")
}

// DeleteProductFromIndex 从 Elasticsearch 索引中删除商品
func (r *productRepository) DeleteProductFromIndex(ctx context.Context, productID uint) error {
	_, err := r.es.Delete().
		Index("products").
		Id(fmt.Sprintf("%d", productID)).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("从 Elasticsearch 删除商品失败: %w", err)
	}
	return nil
}

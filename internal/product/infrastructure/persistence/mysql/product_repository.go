package mysql

import (
	"context"
	"errors" // 导入标准错误处理库。

	"github.com/wyfcoding/ecommerce/internal/product/domain" // 导入商品领域的领域接口和实体。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// ProductRepository 结构体是 ProductRepository 接口的MySQL实现。
type ProductRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewProductRepository 创建并返回一个新的 ProductRepository 实例。
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Save 将商品实体保存到数据库。
// 如果是新商品，则创建；如果商品ID已存在，则更新。
// 此方法在一个事务中创建商品主实体及其关联的SKU列表。
func (r *ProductRepository) Save(ctx context.Context, product *domain.Product) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建或更新商品主实体。
		// GORM的Create在没有主键值时插入，有主键值时更新。
		// 但在这里，我们假设是创建新产品时才调用Save，并且会级联创建SKU。
		// 如果是Update操作，通常会单独处理Product和SKU。
		if err := tx.Create(product).Error; err != nil {
			return err
		}
		// 遍历SKU列表，将每个SKU与商品关联并创建。
		for _, sku := range product.SKUs {
			sku.ProductID = product.ID // 关联商品ID。
			if err := tx.Create(sku).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// FindByID 根据ID从数据库获取商品记录，并预加载其关联的SKU列表。
func (r *ProductRepository) FindByID(ctx context.Context, id uint) (*domain.Product, error) {
	var product domain.Product
	// Preload "SKUs" 确保在获取商品时，同时加载所有关联的SKU。
	if err := r.db.WithContext(ctx).Preload("SKUs").First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &product, nil
}

// FindByName 根据名称从数据库获取商品记录，并预加载其关联的SKU列表。
func (r *ProductRepository) FindByName(ctx context.Context, name string) (*domain.Product, error) {
	var product domain.Product
	// Preload "SKUs" 确保在获取商品时，同时加载所有关联的SKU。
	if err := r.db.WithContext(ctx).Preload("SKUs").Where("name = ?", name).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &product, nil
}

// Update 更新商品实体。
// 此方法更新商品主实体，但SKU的更新需要单独处理或通过其他逻辑。
func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新商品主实体。
		if err := tx.Save(product).Error; err != nil {
			return err
		}
		// 备注：SKU的更新逻辑需要根据业务需求进行更精细的控制。
		// 例如，如果SKU列表发生变化（增删改），需要单独处理SKU表的插入、更新和删除操作。
		// 目前这里只更新Product主实体。
		return nil
	})
}

// Delete 根据ID从数据库删除商品记录。
// GORM默认进行软删除（设置DeletedAt字段）。
func (r *ProductRepository) Delete(ctx context.Context, id uint) error {
	// 软删除商品，不删除关联的SKU，但可以配置GORM进行级联删除。
	return r.db.WithContext(ctx).Delete(&domain.Product{}, id).Error
}

// List 从数据库列出所有商品记录，支持分页，并预加载其关联的SKU列表。
func (r *ProductRepository) List(ctx context.Context, offset, limit int) ([]*domain.Product, int64, error) {
	var products []*domain.Product
	var total int64

	// 统计总记录数。
	if err := r.db.WithContext(ctx).Model(&domain.Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Preload "SKUs" 确保在获取商品列表时，同时加载所有关联的SKU。
	if err := r.db.WithContext(ctx).Preload("SKUs").Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// ListByCategory 从数据库列出指定分类ID下的商品记录，支持分页，并预加载其关联的SKU列表。
func (r *ProductRepository) ListByCategory(ctx context.Context, categoryID uint, offset, limit int) ([]*domain.Product, int64, error) {
	var products []*domain.Product
	var total int64

	// 统计指定分类下的商品总数。
	if err := r.db.WithContext(ctx).Model(&domain.Product{}).Where("category_id = ?", categoryID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Preload "SKUs" 确保在获取商品列表时，同时加载所有关联的SKU。
	if err := r.db.WithContext(ctx).Preload("SKUs").Where("category_id = ?", categoryID).Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// ListByBrand 从数据库列出指定品牌ID下的商品记录，支持分页，并预加载其关联的SKU列表。
func (r *ProductRepository) ListByBrand(ctx context.Context, brandID uint, offset, limit int) ([]*domain.Product, int64, error) {
	var products []*domain.Product
	var total int64

	// 统计指定品牌下的商品总数。
	if err := r.db.WithContext(ctx).Model(&domain.Product{}).Where("brand_id = ?", brandID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Preload "SKUs" 确保在获取商品列表时，同时加载所有关联的SKU。
	if err := r.db.WithContext(ctx).Preload("SKUs").Where("brand_id = ?", brandID).Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

package mysql

import (
	"context"
	"errors" // 导入标准错误处理库。

	"github.com/wyfcoding/ecommerce/internal/product/domain" // 导入商品领域的领域接口和实体。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// SKURepository 结构体是 SKURepository 接口的MySQL实现。
type SKURepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewSKURepository 创建并返回一个新的 SKURepository 实例。
func NewSKURepository(db *gorm.DB) *SKURepository {
	return &SKURepository{db: db}
}

// Save 将SKU实体保存到数据库。
// 如果是新SKU，则创建；如果SKUID已存在，则更新。
func (r *SKURepository) Save(ctx context.Context, sku *domain.SKU) error {
	// GORM的Create在没有主键值时插入，有主键值时更新。
	// 这里使用Create，假定Save是用于创建新SKU。
	// 对于更新，通常使用Update方法。
	return r.db.WithContext(ctx).Create(sku).Error
}

// FindByID 根据ID从数据库获取SKU记录。
func (r *SKURepository) FindByID(ctx context.Context, id uint) (*domain.SKU, error) {
	var sku domain.SKU
	if err := r.db.WithContext(ctx).First(&sku, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &sku, nil
}

// FindByProductID 根据商品ID从数据库获取所有关联的SKU记录。
func (r *SKURepository) FindByProductID(ctx context.Context, productID uint) ([]*domain.SKU, error) {
	var skus []*domain.SKU
	if err := r.db.WithContext(ctx).Where("product_id = ?", productID).Find(&skus).Error; err != nil {
		return nil, err
	}
	return skus, nil
}

// Update 更新SKU实体。
func (r *SKURepository) Update(ctx context.Context, sku *domain.SKU) error {
	// GORM的Save方法会根据主键判断是插入还是更新。
	return r.db.WithContext(ctx).Save(sku).Error
}

// Delete 根据ID从数据库删除SKU记录。
// GORM默认进行软删除。
func (r *SKURepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.SKU{}, id).Error
}

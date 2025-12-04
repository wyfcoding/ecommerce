package mysql

import (
	"context"
	"errors" // 导入标准错误处理库。

	"github.com/wyfcoding/ecommerce/internal/product/domain" // 导入商品领域的领域接口和实体。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// BrandRepository 结构体是 BrandRepository 接口的MySQL实现。
type BrandRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewBrandRepository 创建并返回一个新的 BrandRepository 实例。
func NewBrandRepository(db *gorm.DB) *BrandRepository {
	return &BrandRepository{db: db}
}

// Save 将品牌实体保存到数据库。
// 如果是新品牌，则创建；如果品牌ID已存在，则更新。
func (r *BrandRepository) Save(ctx context.Context, brand *domain.Brand) error {
	// GORM的Create在没有主键值时插入，有主键值时更新。
	// 这里使用Create，假定Save是用于创建新品牌。
	// 对于更新，通常使用Update方法。
	return r.db.WithContext(ctx).Create(brand).Error
}

// FindByID 根据ID从数据库获取品牌记录。
func (r *BrandRepository) FindByID(ctx context.Context, id uint) (*domain.Brand, error) {
	var brand domain.Brand
	if err := r.db.WithContext(ctx).First(&brand, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &brand, nil
}

// FindByName 根据名称从数据库获取品牌记录。
func (r *BrandRepository) FindByName(ctx context.Context, name string) (*domain.Brand, error) {
	var brand domain.Brand
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&brand).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &brand, nil
}

// Update 更新品牌实体。
func (r *BrandRepository) Update(ctx context.Context, brand *domain.Brand) error {
	// GORM的Save方法会根据主键判断是插入还是更新。
	return r.db.WithContext(ctx).Save(brand).Error
}

// Delete 根据ID从数据库删除品牌记录。
// GORM默认进行软删除。
func (r *BrandRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Brand{}, id).Error
}

// List 从数据库列出所有品牌记录。
func (r *BrandRepository) List(ctx context.Context) ([]*domain.Brand, error) {
	var brands []*domain.Brand
	if err := r.db.WithContext(ctx).Find(&brands).Error; err != nil {
		return nil, err
	}
	return brands, nil
}

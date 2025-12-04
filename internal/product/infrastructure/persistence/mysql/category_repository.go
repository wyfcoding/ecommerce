package mysql

import (
	"context"
	"errors" // 导入标准错误处理库。

	"github.com/wyfcoding/ecommerce/internal/product/domain" // 导入商品领域的领域接口和实体。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// CategoryRepository 结构体是 CategoryRepository 接口的MySQL实现。
type CategoryRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewCategoryRepository 创建并返回一个新的 CategoryRepository 实例。
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Save 将分类实体保存到数据库。
// 如果是新分类，则创建；如果分类ID已存在，则更新。
func (r *CategoryRepository) Save(ctx context.Context, category *domain.Category) error {
	// GORM的Create在没有主键值时插入，有主键值时更新。
	// 这里使用Create，假定Save是用于创建新分类。
	// 对于更新，通常使用Update方法。
	return r.db.WithContext(ctx).Create(category).Error
}

// FindByID 根据ID从数据库获取分类记录。
func (r *CategoryRepository) FindByID(ctx context.Context, id uint) (*domain.Category, error) {
	var category domain.Category
	if err := r.db.WithContext(ctx).First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &category, nil
}

// FindByName 根据名称从数据库获取分类记录。
func (r *CategoryRepository) FindByName(ctx context.Context, name string) (*domain.Category, error) {
	var category domain.Category
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &category, nil
}

// Update 更新分类实体。
func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	// GORM的Save方法会根据主键判断是插入还是更新。
	return r.db.WithContext(ctx).Save(category).Error
}

// Delete 根据ID从数据库删除分类记录。
// GORM默认进行软删除。
func (r *CategoryRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Category{}, id).Error
}

// List 从数据库列出所有分类记录。
func (r *CategoryRepository) List(ctx context.Context) ([]*domain.Category, error) {
	var categories []*domain.Category
	if err := r.db.WithContext(ctx).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// FindByParentID 根据父分类ID从数据库获取子分类记录列表。
func (r *CategoryRepository) FindByParentID(ctx context.Context, parentID uint) ([]*domain.Category, error) {
	var categories []*domain.Category
	if err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

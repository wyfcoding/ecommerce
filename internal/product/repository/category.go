package data

import (
	"context"
	"ecommerce/internal/product/biz"
	"ecommerce/internal/product/data/model"
)

type categoryRepo struct {
	*Data
}

// NewCategoryRepo 是 categoryRepo 的构造函数。
func NewCategoryRepo(data *Data) biz.CategoryRepo {
	return &categoryRepo{Data: data}
}

// toBizCategory 将数据库模型 data.Category 转换为业务领域模型 biz.Category。
func (r *categoryRepo) toBizCategory(c *Category) *biz.Category {
	if c == nil {
		return nil
	}
	return &biz.Category{
		ID:        uint64(c.ID),
		ParentID:  c.ParentID,
		Name:      c.Name,
		Level:     c.Level,
		Icon:      &c.Icon,
		SortOrder: &c.SortOrder,
		IsVisible: &c.IsVisible,
	}
}

// CreateCategory 创建一个新的商品分类。
func (r *categoryRepo) CreateCategory(ctx context.Context, c *biz.Category) (*biz.Category, error) {
	// 注意：这里简化了 Level 的计算，实际应根据 ParentID 查询父分类来确定。
	category := &Category{
		ParentID: c.ParentID,
		Name:     c.Name,
		Level:    c.Level,
	}
	if c.Icon != nil {
		category.Icon = *c.Icon
	}
	if c.SortOrder != nil {
		category.SortOrder = *c.SortOrder
	}
	if c.IsVisible != nil {
		category.IsVisible = *c.IsVisible
	}

	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		return nil, err
	}
	return r.toBizCategory(category), nil
}

// UpdateCategory 更新一个已有的商品分类。
func (r *categoryRepo) UpdateCategory(ctx context.Context, c *biz.Category) (*biz.Category, error) {
	var category Category
	if err := r.db.WithContext(ctx).First(&category, c.ID).Error; err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if c.ParentID != 0 {
		updates["parent_id"] = c.ParentID
	}
	if c.Name != "" {
		updates["name"] = c.Name
	}
	if c.Icon != nil {
		updates["icon"] = *c.Icon
	}
	if c.SortOrder != nil {
		updates["sort_order"] = *c.SortOrder
	}
	if c.IsVisible != nil {
		updates["is_visible"] = *c.IsVisible
	}

	if err := r.db.WithContext(ctx).Model(&category).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.toBizCategory(&category), nil
}

// DeleteCategory 删除一个商品分类。
func (r *categoryRepo) DeleteCategory(ctx context.Context, id uint64) error {
	// 在实际业务中，删除分类前需要检查是否有子分类或商品关联，这里简化处理。
	return r.db.WithContext(ctx).Delete(&Category{}, id).Error
}

// GetCategory 获取单个商品分类的详情。
func (r *categoryRepo) GetCategory(ctx context.Context, id uint64) (*biz.Category, error) {
	var category Category
	if err := r.db.WithContext(ctx).First(&category, id).Error; err != nil {
		return nil, err
	}
	return r.toBizCategory(&category), nil
}

// ListCategories 获取商品分类列表。
func (r *categoryRepo) ListCategories(ctx context.Context, parentID uint64) ([]*biz.Category, error) {
	var categories []*model.Category
	if err := r.db.WithContext(ctx).Where(&model.Category{ParentID: parentID}).Find(&categories).Error; err != nil {
		return nil, err
	}

	var bizCategories []*biz.Category
	for _, c := range categories {
		bizCategories = append(bizCategories, r.toBizCategory(c))
	}
	return bizCategories, nil
}

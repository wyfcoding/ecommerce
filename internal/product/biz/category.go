package biz

import (
	"context"
	"errors"
)

// CategoryUsecase 封装了分类相关的业务逻辑。
type CategoryUsecase struct {
	repo CategoryRepo
}

// NewCategoryUsecase 是 CategoryUsecase 的构造函数。
func NewCategoryUsecase(repo CategoryRepo) *CategoryUsecase {
	return &CategoryUsecase{repo: repo}
}

// CreateCategory 负责创建商品分类的业务逻辑。
func (uc *CategoryUsecase) CreateCategory(ctx context.Context, c *Category) (*Category, error) {
	// 核心业务逻辑：根据父ID计算新分类的层级。
	if c.ParentID == 0 {
		c.Level = 1
	} else {
		parent, err := uc.repo.GetCategory(ctx, c.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent category not found: %w", err)
		}
		c.Level = parent.Level + 1
	}
	return uc.repo.CreateCategory(ctx, c)
}

// UpdateCategory 负责更新商品分类的业务逻辑。
func (uc *CategoryUsecase) UpdateCategory(ctx context.Context, c *Category) (*Category, error) {
	// 此处可添加业务校验，例如检查 ParentID 是否会导致循环引用等。
	return uc.repo.UpdateCategory(ctx, c)
}

// DeleteCategory 负责删除商品分类的业务逻辑。
func (uc *CategoryUsecase) DeleteCategory(ctx context.Context, id uint64) error {
	// 此处可添加业务校验，例如检查该分类下是否有子分类或商品。
	return uc.repo.DeleteCategory(ctx, id)
}

// GetCategory 负责获取单个分类的业务逻辑。
func (uc *CategoryUsecase) GetCategory(ctx context.Context, id uint64) (*Category, error) {
	return uc.repo.GetCategory(ctx, id)
}

// ListCategories 负责获取分类列表的业务逻辑。
func (uc *CategoryUsecase) ListCategories(ctx context.Context, parentID uint64) ([]*Category, error) {
	return uc.repo.ListCategories(ctx, parentID)
}

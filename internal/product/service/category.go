package service

import (
	"context"
	"fmt"

	"ecommerce/internal/product/model"
	"ecommerce/internal/product/repository"
)

// CategoryService 封装了分类相关的业务逻辑。
type CategoryService struct {
	repo repository.CategoryRepo
}

// NewCategoryService 是 CategoryService 的构造函数。
func NewCategoryService(repo repository.CategoryRepo) *CategoryService {
	return &CategoryService{repo: repo}
}

// CreateCategory 负责创建商品分类的业务逻辑。
func (s *CategoryService) CreateCategory(ctx context.Context, c *model.Category) (*model.Category, error) {
	// 核心业务逻辑：根据父ID计算新分类的层级。
	if c.ParentID == 0 {
		c.Level = 1
	} else {
		parent, err := s.repo.GetCategory(ctx, c.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent category not found: %w", err)
		}
		c.Level = parent.Level + 1
	}
	return s.repo.CreateCategory(ctx, c)
}

// UpdateCategory 负责更新商品分类的业务逻辑。
func (s *CategoryService) UpdateCategory(ctx context.Context, c *model.Category) (*model.Category, error) {
	// 此处可添加业务校验，例如检查 ParentID 是否会导致循环引用等。
	return s.repo.UpdateCategory(ctx, c)
}

// DeleteCategory 负责删除商品分类的业务逻辑。
func (s *CategoryService) DeleteCategory(ctx context.Context, id uint64) error {
	// 此处可添加业务校验，例如检查该分类下是否有子分类或商品。
	return s.repo.DeleteCategory(ctx, id)
}

// GetCategory 负责获取单个分类的业务逻辑。
func (s *CategoryService) GetCategory(ctx context.Context, id uint64) (*model.Category, error) {
	return s.repo.GetCategory(ctx, id)
}

// ListCategories 负责获取分类列表的业务逻辑。
func (s *CategoryService) ListCategories(ctx context.Context, parentID uint64) ([]*model.Category, error) {
	return s.repo.ListCategories(ctx, parentID)
}
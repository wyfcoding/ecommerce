package application

import (
	"context"
	"errors"

	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/product/domain"
)

type CategoryService struct {
	repo   domain.CategoryRepository
	logger *slog.Logger
}

func NewCategoryService(repo domain.CategoryRepository, logger *slog.Logger) *CategoryService {
	return &CategoryService{
		repo:   repo,
		logger: logger,
	}
}

// CreateCategory 创建分类
func (s *CategoryService) CreateCategory(ctx context.Context, name string, parentID uint64) (*domain.Category, error) {
	category, err := domain.NewCategory(name, uint(parentID))
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create new category entity", "error", err)
		return nil, err
	}

	if err := s.repo.Save(ctx, category); err != nil {
		s.logger.ErrorContext(ctx, "failed to save category", "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "category created successfully", "category_id", category.ID)
	return category, nil
}

// GetCategoryByID 获取分类
func (s *CategoryService) GetCategoryByID(ctx context.Context, id uint64) (*domain.Category, error) {
	return s.repo.FindByID(ctx, uint(id))
}

// UpdateCategory 更新分类
func (s *CategoryService) UpdateCategory(ctx context.Context, id uint64, name *string, parentID *uint64, sort *int) (*domain.Category, error) {
	category, err := s.repo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, errors.New("category not found")
	}

	if name != nil {
		category.Name = *name
	}
	if parentID != nil {
		category.ParentID = uint(*parentID)
	}
	if sort != nil {
		category.Sort = *sort
	}

	if err := s.repo.Update(ctx, category); err != nil {
		s.logger.ErrorContext(ctx, "failed to update category", "category_id", id, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "category updated successfully", "category_id", id)
	return category, nil
}

// DeleteCategory 删除分类
func (s *CategoryService) DeleteCategory(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, uint(id)); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete category", "category_id", id, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "category deleted successfully", "category_id", id)
	return nil
}

// ListCategories 列出分类
func (s *CategoryService) ListCategories(ctx context.Context, parentID uint64) ([]*domain.Category, error) {
	if parentID > 0 {
		return s.repo.FindByParentID(ctx, uint(parentID))
	}
	return s.repo.List(ctx)
}

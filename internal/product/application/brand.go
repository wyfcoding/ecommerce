package application

import (
	"context"
	"errors"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/product/domain"
)

// BrandService 定义了 Brand 相关的服务逻辑。
type BrandService struct {
	repo   domain.BrandRepository
	logger *slog.Logger
}

// NewBrandService 定义了 NewBrand 相关的服务逻辑。
func NewBrandService(repo domain.BrandRepository, logger *slog.Logger) *BrandService {
	return &BrandService{
		repo:   repo,
		logger: logger,
	}
}

// CreateBrand 创建品牌
func (s *BrandService) CreateBrand(ctx context.Context, name, logo string) (*domain.Brand, error) {
	brand, err := domain.NewBrand(name, logo)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create new brand entity", "error", err)
		return nil, err
	}

	if err := s.repo.Save(ctx, brand); err != nil {
		s.logger.ErrorContext(ctx, "failed to save brand", "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "brand created successfully", "brand_id", brand.ID)
	return brand, nil
}

// GetBrandByID 获取品牌
func (s *BrandService) GetBrandByID(ctx context.Context, id uint64) (*domain.Brand, error) {
	return s.repo.FindByID(ctx, uint(id))
}

// UpdateBrand 更新品牌
func (s *BrandService) UpdateBrand(ctx context.Context, id uint64, name, logo *string) (*domain.Brand, error) {
	brand, err := s.repo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	if brand == nil {
		return nil, errors.New("brand not found")
	}

	if name != nil {
		brand.Name = *name
	}
	if logo != nil {
		brand.Logo = *logo
	}

	if err := s.repo.Update(ctx, brand); err != nil {
		s.logger.ErrorContext(ctx, "failed to update brand", "brand_id", id, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "brand updated successfully", "brand_id", id)
	return brand, nil
}

// DeleteBrand 删除品牌
func (s *BrandService) DeleteBrand(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, uint(id)); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete brand", "brand_id", id, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "brand deleted successfully", "brand_id", id)
	return nil
}

// ListBrands 列出品牌
func (s *BrandService) ListBrands(ctx context.Context) ([]*domain.Brand, error) {
	return s.repo.List(ctx)
}

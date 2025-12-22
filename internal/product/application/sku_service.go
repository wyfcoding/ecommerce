package application

import (
	"context"
	"errors"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/product/domain"
)

// SKUService 定义了 SKU 相关的服务逻辑。
type SKUService struct {
	productRepo domain.ProductRepository
	skuRepo     domain.SKURepository
	logger      *slog.Logger
}

// NewSKUService 定义了 NewSKU 相关的服务逻辑。
func NewSKUService(productRepo domain.ProductRepository, skuRepo domain.SKURepository, logger *slog.Logger) *SKUService {
	return &SKUService{
		productRepo: productRepo,
		skuRepo:     skuRepo,
		logger:      logger,
	}
}

// AddSKU 添加SKU
func (s *SKUService) AddSKU(ctx context.Context, productID uint64, name string, price int64, stock int32, image string, specs map[string]string) (*domain.SKU, error) {
	product, err := s.productRepo.FindByID(ctx, uint(productID))
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	sku, err := domain.NewSKU(uint(productID), name, price, stock, image, specs)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create new SKU entity", "error", err)
		return nil, err
	}

	if err := s.skuRepo.Save(ctx, sku); err != nil {
		s.logger.ErrorContext(ctx, "failed to save SKU", "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "SKU added successfully", "sku_id", sku.ID, "product_id", productID)

	return sku, nil
}

// UpdateSKU 更新SKU
func (s *SKUService) UpdateSKU(ctx context.Context, id uint64, price *int64, stock *int32, image *string) (*domain.SKU, error) {
	sku, err := s.skuRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	if sku == nil {
		return nil, errors.New("SKU not found")
	}

	if price != nil {
		sku.Price = *price
	}
	if stock != nil {
		sku.Stock = *stock
	}
	if image != nil {
		sku.Image = *image
	}

	if err := s.skuRepo.Update(ctx, sku); err != nil {
		s.logger.ErrorContext(ctx, "failed to update SKU", "sku_id", id, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "SKU updated successfully", "sku_id", id)
	return sku, nil
}

// DeleteSKU 删除SKU
func (s *SKUService) DeleteSKU(ctx context.Context, id uint64) error {
	if err := s.skuRepo.Delete(ctx, uint(id)); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete SKU", "sku_id", id, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "SKU deleted successfully", "sku_id", id)
	return nil
}

// GetSKUByID 获取SKU
func (s *SKUService) GetSKUByID(ctx context.Context, id uint64) (*domain.SKU, error) {
	return s.skuRepo.FindByID(ctx, uint(id))
}

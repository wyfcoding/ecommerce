package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/wyfcoding/ecommerce/internal/product/domain"
	"github.com/wyfcoding/pkg/algorithm"
	"github.com/wyfcoding/pkg/cache"
	"github.com/wyfcoding/pkg/metrics"
)

// CatalogService 定义了 Catalog 相关的服务逻辑。
type CatalogService struct {
	repo        domain.ProductRepository
	cache       cache.Cache
	logger      *slog.Logger
	cacheHits   *prometheus.CounterVec
	cacheMisses *prometheus.CounterVec
}

// NewCatalogService 定义了 NewCatalog 相关的服务逻辑。
func NewCatalogService(repo domain.ProductRepository, cache cache.Cache, logger *slog.Logger, m *metrics.Metrics) *CatalogService {
	cacheHits := m.NewCounterVec(prometheus.CounterOpts{
		Name: "product_cache_hits_total",
		Help: "商品缓存命中总数",
	}, []string{"layer"})

	cacheMisses := m.NewCounterVec(prometheus.CounterOpts{
		Name: "product_cache_misses_total",
		Help: "商品缓存未命中总数",
	}, []string{})

	return &CatalogService{
		repo:        repo,
		cache:       cache,
		logger:      logger,
		cacheHits:   cacheHits,
		cacheMisses: cacheMisses,
	}
}

// CreateProduct 创建商品
func (s *CatalogService) CreateProduct(ctx context.Context, name, description string, categoryID, brandID uint64, price int64, stock int32) (*domain.Product, error) {
	product, err := domain.NewProduct(name, description, uint(categoryID), uint(brandID), price, stock)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create new product entity", "error", err)
		return nil, err
	}

	if err := s.repo.Save(ctx, product); err != nil {
		s.logger.ErrorContext(ctx, "failed to save product", "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "product created successfully", "product_id", product.ID)

	return product, nil
}

// GetProductByID 获取商品详情
func (s *CatalogService) GetProductByID(ctx context.Context, id uint64) (*domain.Product, error) {
	var product domain.Product
	cacheKey := fmt.Sprintf("product:%d", id)
	if err := s.cache.Get(ctx, cacheKey, &product); err == nil {
		s.cacheHits.WithLabelValues("L1_L2").Inc()
		return &product, nil
	}
	s.cacheMisses.WithLabelValues().Inc()

	p, err := s.repo.FindByID(ctx, uint(id))
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to find product by ID from DB", "product_id", id, "error", err)
		return nil, err
	}
	if p == nil {
		return nil, nil
	}

	_ = s.cache.Set(ctx, cacheKey, p, 10*time.Minute)
	return p, nil
}

// UpdateProductInfo 更新商品信息
func (s *CatalogService) UpdateProductInfo(ctx context.Context, id uint64, name, description *string, categoryID, brandID *uint64, status *domain.ProductStatus) (*domain.Product, error) {
	product, err := s.repo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	if name != nil {
		product.Name = *name
	}
	if description != nil {
		product.Description = *description
	}
	if categoryID != nil {
		product.CategoryID = uint(*categoryID)
	}
	if brandID != nil {
		product.BrandID = uint(*brandID)
	}
	if status != nil {
		product.Status = *status
	}

	if err := s.repo.Update(ctx, product); err != nil {
		s.logger.ErrorContext(ctx, "failed to update product", "product_id", id, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "product updated successfully", "product_id", id)

	_ = s.cache.Delete(ctx, fmt.Sprintf("product:%d", id))

	return product, nil
}

// DeleteProduct 删除商品
func (s *CatalogService) DeleteProduct(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, uint(id)); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete product", "product_id", id, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "product deleted successfully", "product_id", id)
	return s.cache.Delete(ctx, fmt.Sprintf("product:%d", id))
}

// ListProducts 列出商品
func (s *CatalogService) ListProducts(ctx context.Context, page, pageSize int, categoryID, brandID uint64) ([]*domain.Product, int64, error) {
	offset := (page - 1) * pageSize
	if categoryID > 0 {
		return s.repo.ListByCategory(ctx, uint(categoryID), offset, pageSize)
	}
	if brandID > 0 {
		return s.repo.ListByBrand(ctx, uint(brandID), offset, pageSize)
	}
	return s.repo.List(ctx, offset, pageSize)
}

// CalculateProductPrice 计算价格
func (s *CatalogService) CalculateProductPrice(ctx context.Context, productID uint64, userID uint64) (int64, error) {
	product, err := s.repo.FindByID(ctx, uint(productID))
	if err != nil {
		return 0, err
	}
	if product == nil {
		return 0, errors.New("product not found")
	}

	minPrice := int64(float64(product.Price) * 0.8)
	maxPrice := int64(float64(product.Price) * 1.5)
	pe := algorithm.NewPricingEngine(product.Price, minPrice, maxPrice, 1.2)

	factors := algorithm.PricingFactors{
		Stock:           product.Stock,
		TotalStock:      1000,
		DemandLevel:     0.6,
		CompetitorPrice: 0,
		TimeOfDay:       time.Now().Hour(),
		DayOfWeek:       int(time.Now().Weekday()),
		IsHoliday:       false,
		UserLevel:       1,
		SeasonFactor:    1.0,
	}

	if userID > 0 {
		factors.UserLevel = 5
	}

	factors.DemandLevel += (rand.Float64() - 0.5) * 0.2

	return pe.CalculatePrice(factors), nil
}

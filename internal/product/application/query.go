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

type ProductQuery struct {
	repo         domain.ProductRepository
	skuRepo      domain.SKURepository
	brandRepo    domain.BrandRepository
	categoryRepo domain.CategoryRepository
	cache        cache.Cache
	logger       *slog.Logger
	cacheHits    *prometheus.CounterVec
	cacheMisses  *prometheus.CounterVec
}

func NewProductQuery(
	repo domain.ProductRepository,
	skuRepo domain.SKURepository,
	brandRepo domain.BrandRepository,
	categoryRepo domain.CategoryRepository,
	cache cache.Cache,
	logger *slog.Logger,
	m *metrics.Metrics,
) *ProductQuery {
	cacheHits := m.NewCounterVec(prometheus.CounterOpts{
		Name: "product_cache_hits_total",
		Help: "商品缓存命中总数",
	}, []string{"layer"})

	cacheMisses := m.NewCounterVec(prometheus.CounterOpts{
		Name: "product_cache_misses_total",
		Help: "商品缓存未命中总数",
	}, []string{})

	return &ProductQuery{
		repo:         repo,
		skuRepo:      skuRepo,
		brandRepo:    brandRepo,
		categoryRepo: categoryRepo,
		cache:        cache,
		logger:       logger,
		cacheHits:    cacheHits,
		cacheMisses:  cacheMisses,
	}
}

// GetProductByID 获取商品详情
func (q *ProductQuery) GetProductByID(ctx context.Context, id uint64) (*domain.Product, error) {
	var product domain.Product
	cacheKey := fmt.Sprintf("product:%d", id)
	if err := q.cache.Get(ctx, cacheKey, &product); err == nil {
		q.cacheHits.WithLabelValues("L1_L2").Inc()
		return &product, nil
	}
	q.cacheMisses.WithLabelValues().Inc()

	p, err := q.repo.FindByID(ctx, uint(id))
	if err != nil {
		q.logger.ErrorContext(ctx, "failed to find product by ID from DB", "product_id", id, "error", err)
		return nil, err
	}
	if p == nil {
		return nil, nil
	}

	_ = q.cache.Set(ctx, cacheKey, p, 10*time.Minute)
	return p, nil
}

// ListProducts 列出商品
func (q *ProductQuery) ListProducts(ctx context.Context, page, pageSize int, categoryID, brandID uint64) ([]*domain.Product, int64, error) {
	offset := (page - 1) * pageSize
	if categoryID > 0 {
		return q.repo.ListByCategory(ctx, uint(categoryID), offset, pageSize)
	}
	if brandID > 0 {
		return q.repo.ListByBrand(ctx, uint(brandID), offset, pageSize)
	}
	return q.repo.List(ctx, offset, pageSize)
}

// CalculateProductPrice 计算价格
func (q *ProductQuery) CalculateProductPrice(ctx context.Context, productID uint64, userID uint64) (int64, error) {
	product, err := q.repo.FindByID(ctx, uint(productID))
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

// GetSKUByID 获取SKU
func (q *ProductQuery) GetSKUByID(ctx context.Context, id uint64) (*domain.SKU, error) {
	return q.skuRepo.FindByID(ctx, uint(id))
}

// GetBrandByID 获取品牌
func (q *ProductQuery) GetBrandByID(ctx context.Context, id uint64) (*domain.Brand, error) {
	return q.brandRepo.FindByID(ctx, uint(id))
}

// ListBrands 列出品牌
func (q *ProductQuery) ListBrands(ctx context.Context) ([]*domain.Brand, error) {
	return q.brandRepo.List(ctx)
}

// GetCategoryByID 获取分类
func (q *ProductQuery) GetCategoryByID(ctx context.Context, id uint64) (*domain.Category, error) {
	return q.categoryRepo.FindByID(ctx, uint(id))
}

// ListCategories 列出分类
func (q *ProductQuery) ListCategories(ctx context.Context, parentID uint64) ([]*domain.Category, error) {
	if parentID > 0 {
		return q.categoryRepo.FindByParentID(ctx, uint(parentID))
	}
	return q.categoryRepo.List(ctx)
}

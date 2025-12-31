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
	"github.com/wyfcoding/pkg/utils"
	"gorm.io/gorm"
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

// GetProductByID 获取商品详情（集成了缓存、数据库和终极降级逻辑）
func (q *ProductQuery) GetProductByID(ctx context.Context, id uint64) (*domain.Product, error) {
	// 定义主逻辑：缓存 -> DB
	mainFunc := func(c context.Context) (*domain.Product, error) {
		var product domain.Product
		cacheKey := fmt.Sprintf("product:%d", id)

		// 1. 尝试从缓存读取
		if err := q.cache.Get(c, cacheKey, &product); err == nil {
			q.cacheHits.WithLabelValues("L1_L2").Inc()
			return &product, nil
		}
		q.cacheMisses.WithLabelValues().Inc()

		// 2. 缓存未命中，查询 DB
		p, err := q.repo.FindByID(c, uint(id))
		if err != nil {
			return nil, err
		}
		if p == nil {
			return nil, nil // 注意：未找到不属于降级场景
		}

		// 3. 写回缓存 (异步)
		go func() {
			if err := q.cache.Set(context.Background(), cacheKey, p, 10*time.Minute); err != nil {
				q.logger.ErrorContext(context.Background(), "failed to backfill product cache in background", "product_id", id, "error", err)
			}
		}()

		return p, nil
	}

	// 定义终极降级逻辑：DB 宕机或网络中断时的兜底
	fallbackFunc := func(c context.Context) (*domain.Product, error) {
		// 返回一个只包含 ID 和名称的极简对象，状态设为“维护中”
		return &domain.Product{
			Model:       gorm.Model{ID: uint(id)},
			Name:        "商品信息暂时不可用 (服务降级)",
			Description: "目前访问量过大，部分详情暂无法显示，请稍后再试。",
			Price:       0,
			Stock:       0,
		}, nil
	}
	// 使用通用 Fallback 包装器执行
	return utils.ExecuteWithFallback(ctx, "product", "GetProductByID", mainFunc, fallbackFunc)
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

	result := pe.CalculatePrice(factors)
	return result.FinalPrice, nil
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

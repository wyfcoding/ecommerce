// Package application 提供了商品模块的业务逻辑处理。
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
	algorithm "github.com/wyfcoding/pkg/algorithm/finance"
	"github.com/wyfcoding/pkg/cache"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/policy"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

// ProductQuery 处理商品模块的所有只读查询，集成了高性能缓存、并发防击穿及业务降级机制。
type ProductQuery struct {
	repo         domain.ProductRepository  // 商品主仓储
	skuRepo      domain.SKURepository      // SKU 仓储
	brandRepo    domain.BrandRepository    // 品牌仓储
	categoryRepo domain.CategoryRepository // 分类仓储
	cache        cache.Cache               // 缓存组件
	logger       *slog.Logger              // 结构化日志记录器
	cacheHits    *prometheus.CounterVec    // 缓存命中指标统计
	cacheMisses  *prometheus.CounterVec    // 缓存未命中指标统计
	sf           singleflight.Group        // SingleFlight 实例，用于合并瞬时高并发下的同 Key 回源请求
}

// NewProductQuery 初始化并返回一个新的商品查询服务。
func NewProductQuery(
	repo domain.ProductRepository,
	skuRepo domain.SKURepository,
	brandRepo domain.BrandRepository,
	categoryRepo domain.CategoryRepository,
	cache cache.Cache,
	logger *slog.Logger,
	m *metrics.Metrics,
) *ProductQuery {
	cacheHits := m.NewCounterVec(&prometheus.CounterOpts{
		Name: "product_cache_hits_total",
		Help: "商品缓存命中总数",
	}, []string{"layer"})

	cacheMisses := m.NewCounterVec(&prometheus.CounterOpts{
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
		sf:           singleflight.Group{},
	}
}

// GetProductByID 获取商品及其 SKU 详情。
// 架构设计亮点：
// 1. 三级降级：优先读缓存 -> SingleFlight 保护下读数据库 -> 数据库异常时执行业务兜底（Fallback）。
// 2. 指标采集：实时监控各层级的缓存命中率。
func (q *ProductQuery) GetProductByID(ctx context.Context, id uint64) (*domain.Product, error) {
	cacheKey := fmt.Sprintf("product:%d", id)

	mainFunc := func(c context.Context) (*domain.Product, error) {
		var product domain.Product

		if err := q.cache.Get(c, cacheKey, &product); err == nil {
			q.cacheHits.WithLabelValues("L1_L2").Inc()
			return &product, nil
		}
		q.cacheMisses.WithLabelValues().Inc()

		val, err, shared := q.sf.Do(cacheKey, func() (any, error) {
			p, err := q.repo.FindByID(c, uint(id))
			if err != nil {
				return nil, err
			}
			if p == nil {
				return nil, nil
			}

			if err := q.cache.Set(context.Background(), cacheKey, p, 10*time.Minute); err != nil {
				q.logger.ErrorContext(context.Background(), "failed to backfill product cache", "product_id", id, "error", err)
			}
			return p, nil
		})

		if err != nil {
			return nil, err
		}

		if shared {
			q.logger.DebugContext(c, "query result was shared via singleflight", "product_id", id)
		}

		if val == nil {
			return nil, nil
		}

		return val.(*domain.Product), nil
	}

	fallbackFunc := func(c context.Context) (*domain.Product, error) {
		return &domain.Product{
			Model:       gorm.Model{ID: uint(id)},
			Name:        "商品详情暂时不可用",
			Description: "系统繁忙，部分信息展示受限，请稍后再试。",
			Price:       0,
			Stock:       0,
		}, nil
	}

	return policy.ExecuteWithFallback(ctx, "product", "GetProductByID", mainFunc, fallbackFunc)
}

// ListProducts 分页列出商品，支持分类与品牌过滤。
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

// CalculateProductPrice 计算商品动态实时价格。
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

// GetSKUByID 获取 SKU 详情。
func (q *ProductQuery) GetSKUByID(ctx context.Context, id uint64) (*domain.SKU, error) {
	cacheKey := fmt.Sprintf("sku:%d", id)

	val, err, _ := q.sf.Do(cacheKey, func() (any, error) {
		return q.skuRepo.FindByID(ctx, uint(id))
	})
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil
	}
	return val.(*domain.SKU), nil
}

// GetBrandByID 获取品牌详情。
func (q *ProductQuery) GetBrandByID(ctx context.Context, id uint64) (*domain.Brand, error) {
	return q.brandRepo.FindByID(ctx, uint(id))
}

// ListBrands 获取所有品牌。
func (q *ProductQuery) ListBrands(ctx context.Context) ([]*domain.Brand, error) {
	return q.brandRepo.List(ctx)
}

// GetCategoryByID 获取分类详情。
func (q *ProductQuery) GetCategoryByID(ctx context.Context, id uint64) (*domain.Category, error) {
	return q.categoryRepo.FindByID(ctx, uint(id))
}

// ListCategories 获取子分类。
func (q *ProductQuery) ListCategories(ctx context.Context, parentID uint64) ([]*domain.Category, error) {
	if parentID > 0 {
		return q.categoryRepo.FindByParentID(ctx, uint(parentID))
	}
	return q.categoryRepo.List(ctx)
}

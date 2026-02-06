// Package application 提供了商品模块的业务逻辑处理。
package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/wyfcoding/ecommerce/internal/product/domain"
	algorithm "github.com/wyfcoding/pkg/algorithm/finance"
	"github.com/wyfcoding/pkg/policy"
	"golang.org/x/sync/singleflight"
)

// ProductQueryService 处理商品模块的所有只读查询，集成了并发防击穿及业务降级机制。
type ProductQueryService struct {
	repo             domain.ProductRepository
	skuRepo          domain.SKURepository
	brandRepo        domain.BrandRepository
	categoryRepo     domain.CategoryRepository
	readRepo         domain.ProductReadRepository
	skuReadRepo      domain.SKUReadRepository
	brandReadRepo    domain.BrandReadRepository
	categoryReadRepo domain.CategoryReadRepository
	searchRepo       domain.ProductSearchRepository
	logger           *slog.Logger
	sf               singleflight.Group
}

// NewProductQueryService 初始化并返回一个新的商品查询服务。
func NewProductQueryService(
	repo domain.ProductRepository,
	skuRepo domain.SKURepository,
	brandRepo domain.BrandRepository,
	categoryRepo domain.CategoryRepository,
	readRepo domain.ProductReadRepository,
	skuReadRepo domain.SKUReadRepository,
	brandReadRepo domain.BrandReadRepository,
	categoryReadRepo domain.CategoryReadRepository,
	searchRepo domain.ProductSearchRepository,
	logger *slog.Logger,
) *ProductQueryService {
	return &ProductQueryService{
		repo:             repo,
		skuRepo:          skuRepo,
		brandRepo:        brandRepo,
		categoryRepo:     categoryRepo,
		readRepo:         readRepo,
		skuReadRepo:      skuReadRepo,
		brandReadRepo:    brandReadRepo,
		categoryReadRepo: categoryReadRepo,
		searchRepo:       searchRepo,
		logger:           logger,
		sf:               singleflight.Group{},
	}
}

// GetProductByID 获取商品及其 SKU 详情。
// 架构设计亮点：
// 1. 三级降级：优先读缓存 -> SingleFlight 保护下读数据库 -> 数据库异常时执行业务兜底（Fallback）。
// 2. 指标采集：实时监控各层级的缓存命中率。
func (q *ProductQueryService) GetProductByID(ctx context.Context, id uint64) (*domain.Product, error) {
	mainFunc := func(c context.Context) (*domain.Product, error) {
		if q.readRepo != nil {
			if product, err := q.readRepo.GetByID(c, id); err == nil && product != nil {
				return product, nil
			}
		}

		val, err, shared := q.sf.Do(fmt.Sprintf("product:%d", id), func() (any, error) {
			p, err := q.repo.FindByID(c, uint(id))
			if err != nil {
				return nil, err
			}
			if p == nil {
				return nil, nil
			}
			if q.readRepo != nil {
				if err := q.readRepo.Save(context.Background(), p); err != nil && q.logger != nil {
					q.logger.ErrorContext(context.Background(), "failed to backfill product read model", "product_id", id, "error", err)
				}
			}
			return p, nil
		})

		if err != nil {
			return nil, err
		}

		if shared && q.logger != nil {
			q.logger.DebugContext(c, "query result shared via singleflight", "product_id", id)
		}

		if val == nil {
			return nil, nil
		}
		return val.(*domain.Product), nil
	}

	fallbackFunc := func(c context.Context) (*domain.Product, error) {
		return &domain.Product{
			ID:          uint(id),
			Name:        "商品详情暂时不可用",
			Description: "系统繁忙，部分信息展示受限，请稍后再试。",
			Price:       0,
			Stock:       0,
		}, nil
	}

	return policy.ExecuteWithFallback(ctx, "product", "GetProductByID", mainFunc, fallbackFunc)
}

// ListProducts 分页列出商品，支持分类与品牌过滤。
func (q *ProductQueryService) ListProducts(ctx context.Context, page, pageSize int, categoryID, brandID uint64) ([]*domain.Product, int64, error) {
	offset := (page - 1) * pageSize
	if categoryID > 0 {
		return q.repo.ListByCategory(ctx, uint(categoryID), offset, pageSize)
	}
	if brandID > 0 {
		return q.repo.ListByBrand(ctx, uint(brandID), offset, pageSize)
	}
	return q.repo.List(ctx, offset, pageSize)
}

// SearchProducts 利用 Elasticsearch 进行高性能全文检索。
func (q *ProductQueryService) SearchProducts(ctx context.Context, query string, limit int) ([]*domain.Product, error) {
	if q.searchRepo == nil {
		return nil, errors.New("search repository not configured")
	}
	return q.searchRepo.Search(ctx, query, limit)
}

// CalculateProductPrice 计算商品动态实时价格。
func (q *ProductQueryService) CalculateProductPrice(ctx context.Context, productID uint64, userID uint64) (int64, error) {
	product, err := q.GetProductByID(ctx, productID)
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
func (q *ProductQueryService) GetSKUByID(ctx context.Context, id uint64) (*domain.SKU, error) {
	if q.skuReadRepo != nil {
		if sku, err := q.skuReadRepo.GetByID(ctx, id); err == nil && sku != nil {
			return sku, nil
		}
	}

	val, err, _ := q.sf.Do(fmt.Sprintf("sku:%d", id), func() (any, error) {
		p, err := q.skuRepo.FindByID(ctx, uint(id))
		if err != nil {
			return nil, err
		}
		if p == nil {
			return nil, nil
		}

		if q.skuReadRepo != nil {
			_ = q.skuReadRepo.Save(context.Background(), p)
		}
		return p, nil
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
func (q *ProductQueryService) GetBrandByID(ctx context.Context, id uint64) (*domain.Brand, error) {
	if q.brandReadRepo != nil {
		if brand, err := q.brandReadRepo.GetByID(ctx, id); err == nil && brand != nil {
			return brand, nil
		}
	}

	res, err := q.brandRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	if q.brandReadRepo != nil {
		_ = q.brandReadRepo.Save(ctx, res)
	}
	return res, nil
}

// ListBrands 获取所有品牌。
func (q *ProductQueryService) ListBrands(ctx context.Context) ([]*domain.Brand, error) {
	return q.brandRepo.List(ctx)
}

// GetCategoryByID 获取分类详情。
func (q *ProductQueryService) GetCategoryByID(ctx context.Context, id uint64) (*domain.Category, error) {
	if q.categoryReadRepo != nil {
		if cat, err := q.categoryReadRepo.GetByID(ctx, id); err == nil && cat != nil {
			return cat, nil
		}
	}

	res, err := q.categoryRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	if q.categoryReadRepo != nil {
		_ = q.categoryReadRepo.Save(ctx, res)
	}
	return res, nil
}

// ListCategories 获取子分类。
func (q *ProductQueryService) ListCategories(ctx context.Context, parentID uint64) ([]*domain.Category, error) {
	if parentID > 0 {
		return q.categoryRepo.FindByParentID(ctx, uint(parentID))
	}
	return q.categoryRepo.List(ctx)
}

package application

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/wyfcoding/ecommerce/internal/product/domain"
	"github.com/wyfcoding/ecommerce/pkg/algorithm"
	"github.com/wyfcoding/ecommerce/pkg/cache"
	"github.com/wyfcoding/ecommerce/pkg/metrics"
)

// ProductApplicationService 商品应用服务
type ProductApplicationService struct {
	repo         domain.ProductRepository
	skuRepo      domain.SKURepository
	categoryRepo domain.CategoryRepository
	brandRepo    domain.BrandRepository
	cache        cache.Cache
	logger       *slog.Logger

	// Metrics
	cacheHits   *prometheus.CounterVec
	cacheMisses *prometheus.CounterVec
}

// NewProductApplicationService 创建商品应用服务
func NewProductApplicationService(
	repo domain.ProductRepository,
	skuRepo domain.SKURepository,
	categoryRepo domain.CategoryRepository,
	brandRepo domain.BrandRepository,
	cache cache.Cache,
	logger *slog.Logger,
	m *metrics.Metrics,
) *ProductApplicationService {
	cacheHits := m.NewCounterVec(prometheus.CounterOpts{
		Name: "product_cache_hits_total",
		Help: "Total number of cache hits",
	}, []string{"layer"})

	cacheMisses := m.NewCounterVec(prometheus.CounterOpts{
		Name: "product_cache_misses_total",
		Help: "Total number of cache misses",
	}, []string{})

	return &ProductApplicationService{
		repo:         repo,
		skuRepo:      skuRepo,
		categoryRepo: categoryRepo,
		brandRepo:    brandRepo,
		cache:        cache,
		logger:       logger,
		cacheHits:    cacheHits,
		cacheMisses:  cacheMisses,
	}
}

// CreateProduct 创建商品
func (s *ProductApplicationService) CreateProduct(ctx context.Context, name, description string, categoryID, brandID uint64, price int64, stock int32) (*domain.Product, error) {
	product, err := domain.NewProduct(name, description, uint(categoryID), uint(brandID), price, stock)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, product); err != nil {
		s.logger.ErrorContext(ctx, "failed to create product", "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "product created successfully", "product_id", product.ID)

	return product, nil
}

// GetProductByID 获取商品详情
func (s *ProductApplicationService) GetProductByID(ctx context.Context, id uint64) (*domain.Product, error) {
	// 1. Check Cache
	var product domain.Product
	cacheKey := fmt.Sprintf("product:%d", id)
	if err := s.cache.Get(ctx, cacheKey, &product); err == nil {
		s.cacheHits.WithLabelValues("L1_L2").Inc() // MultiLevelCache handles L1/L2 internally, but we count it as a hit here.
		// Ideally MultiLevelCache should report which layer hit, but for now we just count overall hit.
		return &product, nil
	}
	s.cacheMisses.WithLabelValues().Inc()

	// 2. Check DB
	p, err := s.repo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil
	}

	// Set cache
	_ = s.cache.Set(ctx, cacheKey, p, 10*time.Minute)
	return p, nil
}

// UpdateProductInfo 更新商品信息
func (s *ProductApplicationService) UpdateProductInfo(ctx context.Context, id uint64, name, description *string, categoryID, brandID *uint64, status *domain.ProductStatus) (*domain.Product, error) {
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
	// Weight is not in domain model yet, ignoring for now or should add it.
	// Assuming domain model update is out of scope for this step unless critical.
	// Let's stick to existing domain model.

	// product.UpdatedAt = time.Now() // gorm.Model handles this
	if err := s.repo.Update(ctx, product); err != nil {
		s.logger.ErrorContext(ctx, "failed to update product", "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "product updated successfully", "product_id", id)

	// Invalidate cache
	_ = s.cache.Delete(ctx, fmt.Sprintf("product:%d", id))

	return product, nil
}

// DeleteProduct 删除商品
func (s *ProductApplicationService) DeleteProduct(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, uint(id)); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete product", "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "product deleted successfully", "product_id", id)
	return s.cache.Delete(ctx, fmt.Sprintf("product:%d", id))
}

// ListProducts 列出商品
func (s *ProductApplicationService) ListProducts(ctx context.Context, page, pageSize int, categoryID, brandID uint64) ([]*domain.Product, int64, error) {
	offset := (page - 1) * pageSize
	if categoryID > 0 {
		return s.repo.ListByCategory(ctx, uint(categoryID), offset, pageSize)
	}
	if brandID > 0 {
		return s.repo.ListByBrand(ctx, uint(brandID), offset, pageSize)
	}
	return s.repo.List(ctx, offset, pageSize)
}

// AddSKU 添加SKU
func (s *ProductApplicationService) AddSKU(ctx context.Context, productID uint64, name string, price int64, stock int32, image string, specs map[string]string) (*domain.SKU, error) {
	product, err := s.repo.FindByID(ctx, uint(productID))
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	sku, err := domain.NewSKU(uint(productID), name, price, stock, image, specs)
	if err != nil {
		return nil, err
	}

	if err := s.skuRepo.Save(ctx, sku); err != nil {
		s.logger.ErrorContext(ctx, "failed to add SKU", "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "SKU added successfully", "sku_id", sku.ID, "product_id", productID)

	return sku, nil
}

// UpdateSKU 更新SKU
func (s *ProductApplicationService) UpdateSKU(ctx context.Context, id uint64, price *int64, stock *int32, image *string) (*domain.SKU, error) {
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
	// sku.UpdatedAt = time.Now() // gorm.Model handles this

	if err := s.skuRepo.Update(ctx, sku); err != nil {
		s.logger.ErrorContext(ctx, "failed to update SKU", "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "SKU updated successfully", "sku_id", id)
	return sku, nil
}

// DeleteSKU 删除SKU
func (s *ProductApplicationService) DeleteSKU(ctx context.Context, id uint64) error {
	if err := s.skuRepo.Delete(ctx, uint(id)); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete SKU", "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "SKU deleted successfully", "sku_id", id)
	return nil
}

// GetSKUByID 获取SKU
func (s *ProductApplicationService) GetSKUByID(ctx context.Context, id uint64) (*domain.SKU, error) {
	return s.skuRepo.FindByID(ctx, uint(id))
}

// CreateCategory 创建分类
func (s *ProductApplicationService) CreateCategory(ctx context.Context, name string, parentID uint64) (*domain.Category, error) {
	category, err := domain.NewCategory(name, uint(parentID))
	if err != nil {
		return nil, err
	}
	// Add missing fields to domain if needed, or just set what we have
	// Domain Category struct has: ID, Name, ParentID, Sort, Status, CreatedAt, UpdatedAt
	// Missing IconURL in domain struct based on previous view.
	// Let's check domain struct again.
	// It has Sort.
	// category.Sort = sortOrder // Removed sortOrder from input

	if err := s.categoryRepo.Save(ctx, category); err != nil {
		s.logger.ErrorContext(ctx, "failed to create category", "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "category created successfully", "category_id", category.ID)
	return category, nil
}

// GetCategoryByID 获取分类
func (s *ProductApplicationService) GetCategoryByID(ctx context.Context, id uint64) (*domain.Category, error) {
	return s.categoryRepo.FindByID(ctx, uint(id))
}

// UpdateCategory 更新分类
func (s *ProductApplicationService) UpdateCategory(ctx context.Context, id uint64, name *string, parentID *uint64, sort *int) (*domain.Category, error) {
	category, err := s.categoryRepo.FindByID(ctx, uint(id))
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
	// category.UpdatedAt = time.Now() // gorm.Model handles this

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		s.logger.ErrorContext(ctx, "failed to update category", "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "category updated successfully", "category_id", id)
	return category, nil
}

// DeleteCategory 删除分类
func (s *ProductApplicationService) DeleteCategory(ctx context.Context, id uint64) error {
	if err := s.categoryRepo.Delete(ctx, uint(id)); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete category", "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "category deleted successfully", "category_id", id)
	return nil
}

// ListCategories 列出分类
func (s *ProductApplicationService) ListCategories(ctx context.Context, parentID uint64) ([]*domain.Category, error) {
	if parentID > 0 {
		return s.categoryRepo.FindByParentID(ctx, uint(parentID))
	}
	return s.categoryRepo.List(ctx)
}

// CreateBrand 创建品牌
func (s *ProductApplicationService) CreateBrand(ctx context.Context, name, logo string) (*domain.Brand, error) {
	brand, err := domain.NewBrand(name, logo)
	if err != nil {
		return nil, err
	}
	// Description missing in domain Brand struct?
	// Domain Brand: ID, Name, Logo, Status, CreatedAt, UpdatedAt.
	// Ignoring description for now.

	if err := s.brandRepo.Save(ctx, brand); err != nil {
		s.logger.ErrorContext(ctx, "failed to create brand", "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "brand created successfully", "brand_id", brand.ID)
	return brand, nil
}

// GetBrandByID 获取品牌
func (s *ProductApplicationService) GetBrandByID(ctx context.Context, id uint64) (*domain.Brand, error) {
	return s.brandRepo.FindByID(ctx, uint(id))
}

// UpdateBrand 更新品牌
func (s *ProductApplicationService) UpdateBrand(ctx context.Context, id uint64, name, logo *string) (*domain.Brand, error) {
	brand, err := s.brandRepo.FindByID(ctx, uint(id))
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
	// brand.UpdatedAt = time.Now() // gorm.Model handles this

	if err := s.brandRepo.Update(ctx, brand); err != nil {
		s.logger.ErrorContext(ctx, "failed to update brand", "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "brand updated successfully", "brand_id", id)
	return brand, nil
}

// DeleteBrand 删除品牌
func (s *ProductApplicationService) DeleteBrand(ctx context.Context, id uint64) error {
	if err := s.brandRepo.Delete(ctx, uint(id)); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete brand", "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "brand deleted successfully", "brand_id", id)
	return nil
}

// ListBrands 列出品牌
func (s *ProductApplicationService) ListBrands(ctx context.Context) ([]*domain.Brand, error) {
	return s.brandRepo.List(ctx)
}

// CalculateProductPrice 计算商品动态价格
func (s *ProductApplicationService) CalculateProductPrice(ctx context.Context, productID uint64, userID uint64) (int64, error) {
	product, err := s.repo.FindByID(ctx, uint(productID))
	if err != nil {
		return 0, err
	}
	if product == nil {
		return 0, errors.New("product not found")
	}

	// 初始化定价引擎
	// 假设最低价为原价的80%，最高价为原价的150%，弹性系数为1.2
	minPrice := int64(float64(product.Price) * 0.8)
	maxPrice := int64(float64(product.Price) * 1.5)
	pe := algorithm.NewPricingEngine(product.Price, minPrice, maxPrice, 1.2)

	// 构建定价因素
	// 这里使用模拟数据，实际应从各服务获取
	factors := algorithm.PricingFactors{
		Stock:           product.Stock,
		TotalStock:      1000, // 假设总库存
		DemandLevel:     0.6,  // 模拟需求水平
		CompetitorPrice: 0,    // 无竞品数据
		TimeOfDay:       time.Now().Hour(),
		DayOfWeek:       int(time.Now().Weekday()),
		IsHoliday:       false,
		UserLevel:       1, // 默认用户等级
		SeasonFactor:    1.0,
	}

	// 如果有用户ID，可以获取用户等级（这里简化处理）
	if userID > 0 {
		factors.UserLevel = 5 // 假设登录用户等级为5
	}

	// 模拟随机需求波动
	factors.DemandLevel += (rand.Float64() - 0.5) * 0.2

	return pe.CalculatePrice(factors), nil
}

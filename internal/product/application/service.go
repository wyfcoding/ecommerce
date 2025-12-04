package application

import (
	"context"
	"errors"       // 导入标准错误处理库。
	"fmt"          // 导入格式化库。
	"math/rand/v2" // 导入随机数生成库。
	"time"         // 导入时间库。

	"log/slog" // 导入结构化日志库。

	"github.com/prometheus/client_golang/prometheus"         // 导入Prometheus客户端库。
	"github.com/wyfcoding/ecommerce/internal/product/domain" // 导入商品领域的领域接口和实体。
	"github.com/wyfcoding/ecommerce/pkg/algorithm"           // 导入算法包，用于定价引擎。
	"github.com/wyfcoding/ecommerce/pkg/cache"               // 导入缓存库。
	"github.com/wyfcoding/ecommerce/pkg/metrics"             // 导入自定义指标库。
)

// ProductApplicationService 商品应用服务结构体。
// 它协调领域层和基础设施层，处理商品、SKU、分类和品牌的CRUD操作，以及动态价格计算等业务逻辑。
type ProductApplicationService struct {
	repo         domain.ProductRepository  // 商品仓储接口。
	skuRepo      domain.SKURepository      // SKU仓储接口。
	categoryRepo domain.CategoryRepository // 分类仓储接口。
	brandRepo    domain.BrandRepository    // 品牌仓储接口。
	cache        cache.Cache               // 缓存服务接口。
	logger       *slog.Logger              // 日志记录器。

	// Metrics
	cacheHits   *prometheus.CounterVec // 缓存命中计数器。
	cacheMisses *prometheus.CounterVec // 缓存未命中计数器。
}

// NewProductApplicationService 创建商品应用服务实例。
func NewProductApplicationService(
	repo domain.ProductRepository,
	skuRepo domain.SKURepository,
	categoryRepo domain.CategoryRepository,
	brandRepo domain.BrandRepository,
	cache cache.Cache,
	logger *slog.Logger,
	m *metrics.Metrics,
) *ProductApplicationService {
	// 初始化缓存命中计数器。
	cacheHits := m.NewCounterVec(prometheus.CounterOpts{
		Name: "product_cache_hits_total",
		Help: "Total number of cache hits",
	}, []string{"layer"})

	// 初始化缓存未命中计数器。
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

// CreateProduct 创建商品。
func (s *ProductApplicationService) CreateProduct(ctx context.Context, name, description string, categoryID, brandID uint64, price int64, stock int32) (*domain.Product, error) {
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

// GetProductByID 获取商品详情。
// 实现了缓存旁路（Cache-Aside）模式。
func (s *ProductApplicationService) GetProductByID(ctx context.Context, id uint64) (*domain.Product, error) {
	// 1. 尝试从缓存获取商品数据。
	var product domain.Product
	cacheKey := fmt.Sprintf("product:%d", id)
	if err := s.cache.Get(ctx, cacheKey, &product); err == nil {
		s.cacheHits.WithLabelValues("L1_L2").Inc() // 记录缓存命中。
		// MultiLevelCache 在内部处理L1/L2缓存逻辑，此处统一记为命中。
		return &product, nil
	}
	s.cacheMisses.WithLabelValues().Inc() // 记录缓存未命中。

	// 2. 缓存未命中，从数据库获取商品数据。
	p, err := s.repo.FindByID(ctx, uint(id))
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to find product by ID from DB", "product_id", id, "error", err)
		return nil, err
	}
	if p == nil {
		return nil, nil // 商品未找到。
	}

	// 3. 将从数据库获取的数据存入缓存。
	// 缓存有效期设置为10分钟，根据业务需求可调整。
	_ = s.cache.Set(ctx, cacheKey, p, 10*time.Minute)
	return p, nil
}

// UpdateProductInfo 更新商品信息。
func (s *ProductApplicationService) UpdateProductInfo(ctx context.Context, id uint64, name, description *string, categoryID, brandID *uint64, status *domain.ProductStatus) (*domain.Product, error) {
	product, err := s.repo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	// 根据传入的指针参数更新商品信息。
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
	// 备注：领域模型中目前未包含商品重量（Weight）字段。
	// 如果需要更新，应先在领域模型中添加该字段。

	// 更新数据库中的商品信息。
	if err := s.repo.Update(ctx, product); err != nil {
		s.logger.ErrorContext(ctx, "failed to update product", "product_id", id, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "product updated successfully", "product_id", id)

	// 更新操作后，使缓存失效，确保下次查询获取最新数据。
	_ = s.cache.Delete(ctx, fmt.Sprintf("product:%d", id))

	return product, nil
}

// DeleteProduct 删除商品。
func (s *ProductApplicationService) DeleteProduct(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, uint(id)); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete product", "product_id", id, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "product deleted successfully", "product_id", id)
	// 删除商品后，使缓存失效。
	return s.cache.Delete(ctx, fmt.Sprintf("product:%d", id))
}

// ListProducts 列出商品列表。
// 支持按分类ID或品牌ID过滤商品。
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

// AddSKU 添加SKU。
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

// UpdateSKU 更新SKU信息。
func (s *ProductApplicationService) UpdateSKU(ctx context.Context, id uint64, price *int64, stock *int32, image *string) (*domain.SKU, error) {
	sku, err := s.skuRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	if sku == nil {
		return nil, errors.New("SKU not found")
	}

	// 根据传入的指针参数更新SKU信息。
	if price != nil {
		sku.Price = *price
	}
	if stock != nil {
		sku.Stock = *stock
	}
	if image != nil {
		sku.Image = *image
	}
	// 备注：gorm.Model 会自动处理 UpdatedAt 字段。

	if err := s.skuRepo.Update(ctx, sku); err != nil {
		s.logger.ErrorContext(ctx, "failed to update SKU", "sku_id", id, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "SKU updated successfully", "sku_id", id)
	return sku, nil
}

// DeleteSKU 删除SKU。
func (s *ProductApplicationService) DeleteSKU(ctx context.Context, id uint64) error {
	if err := s.skuRepo.Delete(ctx, uint(id)); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete SKU", "sku_id", id, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "SKU deleted successfully", "sku_id", id)
	return nil
}

// GetSKUByID 获取SKU详情。
func (s *ProductApplicationService) GetSKUByID(ctx context.Context, id uint64) (*domain.SKU, error) {
	return s.skuRepo.FindByID(ctx, uint(id))
}

// CreateCategory 创建商品分类。
func (s *ProductApplicationService) CreateCategory(ctx context.Context, name string, parentID uint64) (*domain.Category, error) {
	category, err := domain.NewCategory(name, uint(parentID))
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create new category entity", "error", err)
		return nil, err
	}
	// 备注：领域模型中 Category 结构体目前包含 ID, Name, ParentID, Sort, Status, CreatedAt, UpdatedAt。
	// 缺少 IconURL。Sort 字段在此处未从输入获取。

	if err := s.categoryRepo.Save(ctx, category); err != nil {
		s.logger.ErrorContext(ctx, "failed to save category", "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "category created successfully", "category_id", category.ID)
	return category, nil
}

// GetCategoryByID 获取商品分类详情。
func (s *ProductApplicationService) GetCategoryByID(ctx context.Context, id uint64) (*domain.Category, error) {
	return s.categoryRepo.FindByID(ctx, uint(id))
}

// UpdateCategory 更新商品分类信息。
func (s *ProductApplicationService) UpdateCategory(ctx context.Context, id uint64, name *string, parentID *uint64, sort *int) (*domain.Category, error) {
	category, err := s.categoryRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, errors.New("category not found")
	}

	// 根据传入的指针参数更新分类信息。
	if name != nil {
		category.Name = *name
	}
	if parentID != nil {
		category.ParentID = uint(*parentID)
	}
	if sort != nil {
		category.Sort = *sort
	}
	// 备注：gorm.Model 会自动处理 UpdatedAt 字段。

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		s.logger.ErrorContext(ctx, "failed to update category", "category_id", id, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "category updated successfully", "category_id", id)
	return category, nil
}

// DeleteCategory 删除商品分类。
func (s *ProductApplicationService) DeleteCategory(ctx context.Context, id uint64) error {
	if err := s.categoryRepo.Delete(ctx, uint(id)); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete category", "category_id", id, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "category deleted successfully", "category_id", id)
	return nil
}

// ListCategories 列出商品分类列表。
// 支持按父分类ID过滤。
func (s *ProductApplicationService) ListCategories(ctx context.Context, parentID uint64) ([]*domain.Category, error) {
	if parentID > 0 {
		return s.categoryRepo.FindByParentID(ctx, uint(parentID))
	}
	return s.categoryRepo.List(ctx)
}

// CreateBrand 创建商品品牌。
func (s *ProductApplicationService) CreateBrand(ctx context.Context, name, logo string) (*domain.Brand, error) {
	brand, err := domain.NewBrand(name, logo)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create new brand entity", "error", err)
		return nil, err
	}
	// 备注：领域模型中 Brand 结构体包含 ID, Name, Logo, Status, CreatedAt, UpdatedAt。
	// 缺少 Description。此处未从输入获取 Description。

	if err := s.brandRepo.Save(ctx, brand); err != nil {
		s.logger.ErrorContext(ctx, "failed to save brand", "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "brand created successfully", "brand_id", brand.ID)
	return brand, nil
}

// GetBrandByID 获取商品品牌详情。
func (s *ProductApplicationService) GetBrandByID(ctx context.Context, id uint64) (*domain.Brand, error) {
	return s.brandRepo.FindByID(ctx, uint(id))
}

// UpdateBrand 更新商品品牌信息。
func (s *ProductApplicationService) UpdateBrand(ctx context.Context, id uint64, name, logo *string) (*domain.Brand, error) {
	brand, err := s.brandRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	if brand == nil {
		return nil, errors.New("brand not found")
	}

	// 根据传入的指针参数更新品牌信息。
	if name != nil {
		brand.Name = *name
	}
	if logo != nil {
		brand.Logo = *logo
	}
	// 备注：gorm.Model 会自动处理 UpdatedAt 字段。

	if err := s.brandRepo.Update(ctx, brand); err != nil {
		s.logger.ErrorContext(ctx, "failed to update brand", "brand_id", id, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "brand updated successfully", "brand_id", id)
	return brand, nil
}

// DeleteBrand 删除商品品牌。
func (s *ProductApplicationService) DeleteBrand(ctx context.Context, id uint64) error {
	if err := s.brandRepo.Delete(ctx, uint(id)); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete brand", "brand_id", id, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "brand deleted successfully", "brand_id", id)
	return nil
}

// ListBrands 列出商品品牌列表。
func (s *ProductApplicationService) ListBrands(ctx context.Context) ([]*domain.Brand, error) {
	return s.brandRepo.List(ctx)
}

// CalculateProductPrice 计算商品的动态价格。
// 此方法根据商品ID和用户ID，结合动态定价算法计算出推荐价格。
// ctx: 上下文。
// productID: 商品ID。
// userID: 用户ID（用于个性化定价）。
// 返回计算后的价格（单位：分）和可能发生的错误。
func (s *ProductApplicationService) CalculateProductPrice(ctx context.Context, productID uint64, userID uint64) (int64, error) {
	// 获取商品基本信息。
	product, err := s.repo.FindByID(ctx, uint(productID))
	if err != nil {
		return 0, err
	}
	if product == nil {
		return 0, errors.New("product not found")
	}

	// 1. 初始化定价引擎。
	// 假设最低价为原价的80%，最高价为原价的150%，弹性系数为1.2。
	minPrice := int64(float64(product.Price) * 0.8)
	maxPrice := int64(float64(product.Price) * 1.5)
	pe := algorithm.NewPricingEngine(product.Price, minPrice, maxPrice, 1.2)

	// 2. 构建定价因素（PricingFactors）。
	// 这些因素模拟了影响价格的各种市场和用户行为数据。
	factors := algorithm.PricingFactors{
		Stock:           product.Stock,
		TotalStock:      1000, // 假设总库存为1000，实际应从库存服务获取。
		DemandLevel:     0.6,  // 模拟需求水平，实际应从市场分析服务获取。
		CompetitorPrice: 0,    // 模拟竞争对手价格，实际应从竞品分析服务获取。
		TimeOfDay:       time.Now().Hour(),
		DayOfWeek:       int(time.Now().Weekday()),
		IsHoliday:       false,
		UserLevel:       1, // 默认用户等级为1。
		SeasonFactor:    1.0,
	}

	// 3. 根据用户ID获取用户等级（如果存在）。
	if userID > 0 {
		// TODO: 实际生产中，应调用用户服务获取用户等级信息。
		factors.UserLevel = 5 // 假设登录用户等级为5（模拟数据）。
	}

	// 4. 模拟随机需求波动，增加价格的动态性。
	factors.DemandLevel += (rand.Float64() - 0.5) * 0.2 // 在 -0.1 到 +0.1 之间波动。

	// 5. 调用定价引擎计算最终价格。
	return pe.CalculatePrice(factors), nil
}

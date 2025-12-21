package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/product/domain"
)

// ProductService 商品服务门面（Facade）
// 聚合了 Catalog, Category, Brand, SKU 等子服务
type ProductService struct {
	Catalog  *CatalogService
	Category *CategoryService
	Brand    *BrandService
	SKU      *SKUService
	logger   *slog.Logger
}

// NewProductService 创建商品服务实例。
func NewProductService(
	catalog *CatalogService,
	category *CategoryService,
	brand *BrandService,
	sku *SKUService,
	logger *slog.Logger,
) *ProductService {
	return &ProductService{
		Catalog:  catalog,
		Category: category,
		Brand:    brand,
		SKU:      sku,
		logger:   logger,
	}
}

// --- 委托给子服务的方法 ---
// 保留这些方法以兼容现有接口调用

// Catalog 委托方法

// CreateProduct 创建商品。
func (s *ProductService) CreateProduct(ctx context.Context, name, description string, categoryID, brandID uint64, price int64, stock int32) (*domain.Product, error) {
	return s.Catalog.CreateProduct(ctx, name, description, categoryID, brandID, price, stock)
}

// GetProductByID 根据ID获取商品。
func (s *ProductService) GetProductByID(ctx context.Context, id uint64) (*domain.Product, error) {
	return s.Catalog.GetProductByID(ctx, id)
}

// UpdateProductInfo 更新商品信息。
func (s *ProductService) UpdateProductInfo(ctx context.Context, id uint64, name, description *string, categoryID, brandID *uint64, status *domain.ProductStatus) (*domain.Product, error) {
	return s.Catalog.UpdateProductInfo(ctx, id, name, description, categoryID, brandID, status)
}

// DeleteProduct 删除商品。
func (s *ProductService) DeleteProduct(ctx context.Context, id uint64) error {
	return s.Catalog.DeleteProduct(ctx, id)
}

// ListProducts 获取商品列表。
func (s *ProductService) ListProducts(ctx context.Context, page, pageSize int, categoryID, brandID uint64) ([]*domain.Product, int64, error) {
	return s.Catalog.ListProducts(ctx, page, pageSize, categoryID, brandID)
}

// CalculateProductPrice 计算商品价格。
func (s *ProductService) CalculateProductPrice(ctx context.Context, productID uint64, userID uint64) (int64, error) {
	return s.Catalog.CalculateProductPrice(ctx, productID, userID)
}

// 分类委托方法

// CreateCategory 创建分类。
func (s *ProductService) CreateCategory(ctx context.Context, name string, parentID uint64) (*domain.Category, error) {
	return s.Category.CreateCategory(ctx, name, parentID)
}

// GetCategoryByID 根据ID获取分类。
func (s *ProductService) GetCategoryByID(ctx context.Context, id uint64) (*domain.Category, error) {
	return s.Category.GetCategoryByID(ctx, id)
}

// UpdateCategory 更新分类信息。
func (s *ProductService) UpdateCategory(ctx context.Context, id uint64, name *string, parentID *uint64, sort *int) (*domain.Category, error) {
	return s.Category.UpdateCategory(ctx, id, name, parentID, sort)
}

// DeleteCategory 删除分类。
func (s *ProductService) DeleteCategory(ctx context.Context, id uint64) error {
	return s.Category.DeleteCategory(ctx, id)
}

// ListCategories 获取分类列表。
func (s *ProductService) ListCategories(ctx context.Context, parentID uint64) ([]*domain.Category, error) {
	return s.Category.ListCategories(ctx, parentID)
}

// 品牌委托方法

// CreateBrand 创建品牌。
func (s *ProductService) CreateBrand(ctx context.Context, name, logo string) (*domain.Brand, error) {
	return s.Brand.CreateBrand(ctx, name, logo)
}

// GetBrandByID 根据ID获取品牌。
func (s *ProductService) GetBrandByID(ctx context.Context, id uint64) (*domain.Brand, error) {
	return s.Brand.GetBrandByID(ctx, id)
}

// UpdateBrand 更新品牌信息。
func (s *ProductService) UpdateBrand(ctx context.Context, id uint64, name, logo *string) (*domain.Brand, error) {
	return s.Brand.UpdateBrand(ctx, id, name, logo)
}

// DeleteBrand 删除品牌。
func (s *ProductService) DeleteBrand(ctx context.Context, id uint64) error {
	return s.Brand.DeleteBrand(ctx, id)
}

// ListBrands 获取品牌列表。
func (s *ProductService) ListBrands(ctx context.Context) ([]*domain.Brand, error) {
	return s.Brand.ListBrands(ctx)
}

// AddSKU 添加 SKU。
func (s *ProductService) AddSKU(ctx context.Context, productID uint64, name string, price int64, stock int32, image string, specs map[string]string) (*domain.SKU, error) {
	return s.SKU.AddSKU(ctx, productID, name, price, stock, image, specs)
}

// UpdateSKU 更新 SKU 信息。
func (s *ProductService) UpdateSKU(ctx context.Context, id uint64, price *int64, stock *int32, image *string) (*domain.SKU, error) {
	return s.SKU.UpdateSKU(ctx, id, price, stock, image)
}

// DeleteSKU 删除 SKU。
func (s *ProductService) DeleteSKU(ctx context.Context, id uint64) error {
	return s.SKU.DeleteSKU(ctx, id)
}

// GetSKUByID 根据ID获取 SKU。
func (s *ProductService) GetSKUByID(ctx context.Context, id uint64) (*domain.SKU, error) {
	return s.SKU.GetSKUByID(ctx, id)
}

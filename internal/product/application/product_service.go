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

// --- Delegate Methods to Sub-Services ---
// 保留这些方法以兼容现有接口调用

// Catalog Delegates
func (s *ProductService) CreateProduct(ctx context.Context, name, description string, categoryID, brandID uint64, price int64, stock int32) (*domain.Product, error) {
	return s.Catalog.CreateProduct(ctx, name, description, categoryID, brandID, price, stock)
}

func (s *ProductService) GetProductByID(ctx context.Context, id uint64) (*domain.Product, error) {
	return s.Catalog.GetProductByID(ctx, id)
}

func (s *ProductService) UpdateProductInfo(ctx context.Context, id uint64, name, description *string, categoryID, brandID *uint64, status *domain.ProductStatus) (*domain.Product, error) {
	return s.Catalog.UpdateProductInfo(ctx, id, name, description, categoryID, brandID, status)
}

func (s *ProductService) DeleteProduct(ctx context.Context, id uint64) error {
	return s.Catalog.DeleteProduct(ctx, id)
}

func (s *ProductService) ListProducts(ctx context.Context, page, pageSize int, categoryID, brandID uint64) ([]*domain.Product, int64, error) {
	return s.Catalog.ListProducts(ctx, page, pageSize, categoryID, brandID)
}

func (s *ProductService) CalculateProductPrice(ctx context.Context, productID uint64, userID uint64) (int64, error) {
	return s.Catalog.CalculateProductPrice(ctx, productID, userID)
}

// Category Delegates
func (s *ProductService) CreateCategory(ctx context.Context, name string, parentID uint64) (*domain.Category, error) {
	return s.Category.CreateCategory(ctx, name, parentID)
}

func (s *ProductService) GetCategoryByID(ctx context.Context, id uint64) (*domain.Category, error) {
	return s.Category.GetCategoryByID(ctx, id)
}

func (s *ProductService) UpdateCategory(ctx context.Context, id uint64, name *string, parentID *uint64, sort *int) (*domain.Category, error) {
	return s.Category.UpdateCategory(ctx, id, name, parentID, sort)
}

func (s *ProductService) DeleteCategory(ctx context.Context, id uint64) error {
	return s.Category.DeleteCategory(ctx, id)
}

func (s *ProductService) ListCategories(ctx context.Context, parentID uint64) ([]*domain.Category, error) {
	return s.Category.ListCategories(ctx, parentID)
}

// Brand Delegates
func (s *ProductService) CreateBrand(ctx context.Context, name, logo string) (*domain.Brand, error) {
	return s.Brand.CreateBrand(ctx, name, logo)
}

func (s *ProductService) GetBrandByID(ctx context.Context, id uint64) (*domain.Brand, error) {
	return s.Brand.GetBrandByID(ctx, id)
}

func (s *ProductService) UpdateBrand(ctx context.Context, id uint64, name, logo *string) (*domain.Brand, error) {
	return s.Brand.UpdateBrand(ctx, id, name, logo)
}

func (s *ProductService) DeleteBrand(ctx context.Context, id uint64) error {
	return s.Brand.DeleteBrand(ctx, id)
}

func (s *ProductService) ListBrands(ctx context.Context) ([]*domain.Brand, error) {
	return s.Brand.ListBrands(ctx)
}

// SKU Delegates
func (s *ProductService) AddSKU(ctx context.Context, productID uint64, name string, price int64, stock int32, image string, specs map[string]string) (*domain.SKU, error) {
	return s.SKU.AddSKU(ctx, productID, name, price, stock, image, specs)
}

func (s *ProductService) UpdateSKU(ctx context.Context, id uint64, price *int64, stock *int32, image *string) (*domain.SKU, error) {
	return s.SKU.UpdateSKU(ctx, id, price, stock, image)
}

func (s *ProductService) DeleteSKU(ctx context.Context, id uint64) error {
	return s.SKU.DeleteSKU(ctx, id)
}

func (s *ProductService) GetSKUByID(ctx context.Context, id uint64) (*domain.SKU, error) {
	return s.SKU.GetSKUByID(ctx, id)
}

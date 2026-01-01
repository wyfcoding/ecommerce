package application

import (
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/product/domain"
	"github.com/wyfcoding/pkg/cache"
	"github.com/wyfcoding/pkg/metrics"
)

// ProductService 商品服务门面
type ProductService struct {
	Manager *ProductManager
	Query   *ProductQuery
	logger  *slog.Logger
}

func NewProductService(
	repo domain.ProductRepository,
	skuRepo domain.SKURepository, // Note: Consolidated into ProductRepository struct, but interface might be separate?
	// Actually in my consolidated product_repository.go, I kept separate NewXXXRepository functions but they are in the same file.
	// Wait, I should check if I need to update main.go to inject the same instance or different instances.
	// For simplicity in Application layer, I will inject the repositories as interfaces.
	// I Consolidated the IMPLEMENTATION, but interfaces in `domain/product_repository.go` are still separate.
	// So I still need to pass them separately or aggregate them.
	// Let's pass them separately as defined in domain.
	brandRepo domain.BrandRepository,
	categoryRepo domain.CategoryRepository,
	cache cache.Cache,
	logger *slog.Logger,
	m *metrics.Metrics,
) *ProductService {
	return &ProductService{
		Manager: NewProductManager(repo, skuRepo, brandRepo, categoryRepo, cache, logger),
		Query:   NewProductQuery(repo, skuRepo, brandRepo, categoryRepo, cache, logger, m),
		logger:  logger,
	}
}

// DTOs

type CreateProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CategoryID  uint64 `json:"category_id"`
	BrandID     uint64 `json:"brand_id"`
	Price       int64  `json:"price"`
	Stock       int32  `json:"stock"`
}

type UpdateProductRequest struct {
	Name        *string               `json:"name"`
	Description *string               `json:"description"`
	CategoryID  *uint64               `json:"category_id"`
	BrandID     *uint64               `json:"brand_id"`
	Status      *domain.ProductStatus `json:"status"`
}

type AddSKURequest struct {
	Name  string            `json:"name"`
	Price int64             `json:"price"`
	Stock int32             `json:"stock"`
	Image string            `json:"image"`
	Specs map[string]string `json:"specs"`
}

type UpdateSKURequest struct {
	Price *int64  `json:"price"`
	Stock *int32  `json:"stock"`
	Image *string `json:"image"`
}

type CreateBrandRequest struct {
	Name string `json:"name"`
	Logo string `json:"logo"`
}

type UpdateBrandRequest struct {
	Name *string `json:"name"`
	Logo *string `json:"logo"`
}

type CreateCategoryRequest struct {
	Name     string `json:"name"`
	ParentID uint64 `json:"parent_id"`
}

type UpdateCategoryRequest struct {
	Name     *string `json:"name"`
	ParentID *uint64 `json:"parent_id"`
	Sort     *int    `json:"sort"`
}

// Response mappings can be handled in interface layer or here.
// Returning domain entities is fine for now as per previous service patterns.

package application

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/product/domain"
)

type ProductDTO struct {
	ID          uint                 `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	CategoryID  uint                 `json:"category_id"`
	BrandID     uint                 `json:"brand_id"`
	Price       int64                `json:"price"`
	Stock       int32                `json:"stock"`
	Status      domain.ProductStatus `json:"status"`
	Type        domain.ProductType   `json:"type"`
	MainImage   string               `json:"main_image"`
	Images      []string             `json:"images"`
	SKUs        []*SKUDTO            `json:"skus"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type SKUDTO struct {
	ID        uint              `json:"id"`
	ProductID uint              `json:"product_id"`
	Name      string            `json:"name"`
	Price     int64             `json:"price"`
	Stock     int32             `json:"stock"`
	Image     string            `json:"image"`
	Specs     map[string]string `json:"specs"`
}

// --- Product Requests ---

type CreateProductRequest struct {
	Name        string             `json:"name" binding:"required"`
	Description string             `json:"description"`
	CategoryID  uint               `json:"category_id" binding:"required"`
	BrandID     uint               `json:"brand_id" binding:"required"`
	Type        domain.ProductType `json:"type"`
	Price       int64              `json:"price" binding:"required"`
	Stock       int32              `json:"stock" binding:"required"`
}

type UpdateProductRequest struct {
	Name        *string               `json:"name"`
	Description *string               `json:"description"`
	CategoryID  *uint                 `json:"category_id"`
	BrandID     *uint                 `json:"brand_id"`
	Status      *domain.ProductStatus `json:"status"`
}

// --- SKU Requests ---

type AddSKURequest struct {
	Name  string            `json:"name" binding:"required"`
	Price int64             `json:"price" binding:"required"`
	Stock int32             `json:"stock" binding:"required"`
	Image string            `json:"image"`
	Specs map[string]string `json:"specs"`
}

type UpdateSKURequest struct {
	Price *int64  `json:"price"`
	Stock *int32  `json:"stock"`
	Image *string `json:"image"`
}

// --- Category Requests ---

type CreateCategoryRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID uint   `json:"parent_id"`
}

type UpdateCategoryRequest struct {
	Name     *string `json:"name"`
	ParentID *uint   `json:"parent_id"`
	Sort     *int    `json:"sort"`
}

// --- Brand Requests ---

type CreateBrandRequest struct {
	Name string `json:"name" binding:"required"`
	Logo string `json:"logo"`
}

type UpdateBrandRequest struct {
	Name *string `json:"name"`
	Logo *string `json:"logo"`
}

// --- Internal Commands ---
// NOTE: I am not using embedding here to keep gRPC handler's struct literals working.

type CreateProductCommand struct {
	Name        string
	Description string
	CategoryID  uint
	BrandID     uint
	Type        domain.ProductType
	Price       int64
	Stock       int32
}

type UpdateProductCommand struct {
	ID          uint
	Name        *string
	Description *string
	CategoryID  *uint
	BrandID     *uint
	Status      *domain.ProductStatus
}

type AddSKUCommand struct {
	ProductID uint
	Name      string
	Price     int64
	Stock     int32
	Image     string
	Specs     map[string]string
}

type UpdateSKUCommand struct {
	ID    uint
	Price *int64
	Stock *int32
	Image *string
}

type CreateCategoryCommand struct {
	Name     string
	ParentID uint
}

type UpdateCategoryCommand struct {
	ID       uint
	Name     *string
	ParentID *uint
	Sort     *int
}

type CreateBrandCommand struct {
	Name string
	Logo string
}

type UpdateBrandCommand struct {
	ID   uint
	Name *string
	Logo *string
}

package http

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/product/application"
	"github.com/wyfcoding/ecommerce/internal/product/domain"
	"github.com/wyfcoding/pkg/response"
)

// Handler 处理 HTTP 或 gRPC 请求。
type Handler struct {
	app    *application.Product
	logger *slog.Logger
}

// NewHandler 处理 HTTP 或 gRPC 请求。
func NewHandler(app *application.Product, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	products := router.Group("/products")
	{
		products.POST("", h.CreateProduct)
		products.GET("/:id", h.GetProductByID)
		products.PUT("/:id", h.UpdateProductInfo)
		products.DELETE("/:id", h.DeleteProduct)
		products.GET("", h.ListProducts)
		products.POST("/:id/skus", h.AddSKU)
		products.GET("/:id/price", h.CalculateProductPrice)
	}

	skus := router.Group("/skus")
	{
		skus.PUT("/:id", h.UpdateSKU)
		skus.DELETE("/:id", h.DeleteSKU)
		skus.GET("/:id", h.GetSKUByID)
	}

	categories := router.Group("/categories")
	{
		categories.POST("", h.CreateCategory)
		categories.GET("/:id", h.GetCategoryByID)
		categories.PUT("/:id", h.UpdateCategory)
		categories.DELETE("/:id", h.DeleteCategory)
		categories.GET("", h.ListCategories)
	}

	brands := router.Group("/brands")
	{
		brands.POST("", h.CreateBrand)
		brands.GET("/:id", h.GetBrandByID)
		brands.PUT("/:id", h.UpdateBrand)
		brands.DELETE("/:id", h.DeleteBrand)
		brands.GET("", h.ListBrands)
	}
}

// --- Product Handlers ---

type createProductRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	CategoryID  uint64 `json:"category_id" binding:"required"`
	BrandID     uint64 `json:"brand_id" binding:"required"`
	Price       int64  `json:"price" binding:"required,gt=0"`
	Stock       int32  `json:"stock" binding:"required,gte=0"`
}

func (h *Handler) CreateProduct(c *gin.Context) {
	var req createProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	product, err := h.app.CreateProduct(c.Request.Context(), req.Name, req.Description, req.CategoryID, req.BrandID, req.Price, req.Stock)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to create product", "error", err)
		response.InternalError(c, "failed to create product: "+err.Error())
		return
	}

	response.Success(c, product)
}

func (h *Handler) GetProductByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid product id: "+err.Error())
		return
	}

	product, err := h.app.GetProductByID(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get product", "id", id, "error", err)
		response.InternalError(c, "failed to get product: "+err.Error())
		return
	}
	if product == nil {
		response.NotFound(c, "product not found")
		return
	}

	response.Success(c, product)
}

type updateProductRequest struct {
	Name        *string               `json:"name"`
	Description *string               `json:"description"`
	CategoryID  *uint64               `json:"category_id"`
	BrandID     *uint64               `json:"brand_id"`
	Status      *domain.ProductStatus `json:"status"`
}

func (h *Handler) UpdateProductInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid product id: "+err.Error())
		return
	}

	var req updateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	product, err := h.app.UpdateProductInfo(c.Request.Context(), id, req.Name, req.Description, req.CategoryID, req.BrandID, req.Status)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to update product", "id", id, "error", err)
		response.InternalError(c, "failed to update product: "+err.Error())
		return
	}

	response.Success(c, product)
}

func (h *Handler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid product id: "+err.Error())
		return
	}

	if err := h.app.DeleteProduct(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to delete product", "id", id, "error", err)
		response.InternalError(c, "failed to delete product: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	categoryID, _ := strconv.ParseUint(c.DefaultQuery("category_id", "0"), 10, 64)
	brandID, _ := strconv.ParseUint(c.DefaultQuery("brand_id", "0"), 10, 64)

	products, total, err := h.app.ListProducts(c.Request.Context(), page, pageSize, categoryID, brandID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list products", "error", err)
		response.InternalError(c, "failed to list products: "+err.Error())
		return
	}

	response.SuccessWithPagination(c, products, total, int32(page), int32(pageSize))
}

func (h *Handler) CalculateProductPrice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid product id: "+err.Error())
		return
	}

	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)

	price, err := h.app.CalculateProductPrice(c.Request.Context(), id, userID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to calculate price", "id", id, "error", err)
		response.InternalError(c, "failed to calculate price: "+err.Error())
		return
	}

	response.Success(c, gin.H{"price": price})
}

// --- SKU Handlers ---

type addSKURequest struct {
	Name  string            `json:"name" binding:"required"`
	Price int64             `json:"price" binding:"required,gt=0"`
	Stock int32             `json:"stock" binding:"required,gte=0"`
	Image string            `json:"image"`
	Specs map[string]string `json:"specs"`
}

func (h *Handler) AddSKU(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid product id: "+err.Error())
		return
	}

	var req addSKURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	sku, err := h.app.AddSKU(c.Request.Context(), productID, req.Name, req.Price, req.Stock, req.Image, req.Specs)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to add SKU", "product_id", productID, "error", err)
		response.InternalError(c, "failed to add SKU: "+err.Error())
		return
	}

	response.Success(c, sku)
}

type updateSKURequest struct {
	Price *int64  `json:"price"`
	Stock *int32  `json:"stock"`
	Image *string `json:"image"`
}

func (h *Handler) UpdateSKU(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid sku id: "+err.Error())
		return
	}

	var req updateSKURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	sku, err := h.app.UpdateSKU(c.Request.Context(), id, req.Price, req.Stock, req.Image)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to update SKU", "id", id, "error", err)
		response.InternalError(c, "failed to update SKU: "+err.Error())
		return
	}

	response.Success(c, sku)
}

func (h *Handler) DeleteSKU(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid sku id: "+err.Error())
		return
	}

	if err := h.app.DeleteSKU(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to delete SKU", "id", id, "error", err)
		response.InternalError(c, "failed to delete SKU: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) GetSKUByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid sku id: "+err.Error())
		return
	}

	sku, err := h.app.GetSKUByID(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get SKU", "id", id, "error", err)
		response.InternalError(c, "failed to get SKU: "+err.Error())
		return
	}
	if sku == nil {
		response.NotFound(c, "SKU not found")
		return
	}

	response.Success(c, sku)
}

// --- Category Handlers ---

type createCategoryRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID uint64 `json:"parent_id"`
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var req createCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	category, err := h.app.CreateCategory(c.Request.Context(), req.Name, req.ParentID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to create category", "error", err)
		response.InternalError(c, "failed to create category: "+err.Error())
		return
	}

	response.Success(c, category)
}

func (h *Handler) GetCategoryByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid category id: "+err.Error())
		return
	}

	category, err := h.app.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get category", "id", id, "error", err)
		response.InternalError(c, "failed to get category: "+err.Error())
		return
	}
	if category == nil {
		response.NotFound(c, "category not found")
		return
	}

	response.Success(c, category)
}

type updateCategoryRequest struct {
	Name     *string `json:"name"`
	ParentID *uint64 `json:"parent_id"`
	Sort     *int    `json:"sort"`
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid category id: "+err.Error())
		return
	}

	var req updateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	category, err := h.app.UpdateCategory(c.Request.Context(), id, req.Name, req.ParentID, req.Sort)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to update category", "id", id, "error", err)
		response.InternalError(c, "failed to update category: "+err.Error())
		return
	}

	response.Success(c, category)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid category id: "+err.Error())
		return
	}

	if err := h.app.DeleteCategory(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to delete category", "id", id, "error", err)
		response.InternalError(c, "failed to delete category: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) ListCategories(c *gin.Context) {
	parentID, _ := strconv.ParseUint(c.DefaultQuery("parent_id", "0"), 10, 64)

	categories, err := h.app.ListCategories(c.Request.Context(), parentID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list categories", "error", err)
		response.InternalError(c, "failed to list categories: "+err.Error())
		return
	}

	response.Success(c, categories)
}

// --- Brand Handlers ---

type createBrandRequest struct {
	Name string `json:"name" binding:"required"`
	Logo string `json:"logo"`
}

func (h *Handler) CreateBrand(c *gin.Context) {
	var req createBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	brand, err := h.app.CreateBrand(c.Request.Context(), req.Name, req.Logo)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to create brand", "error", err)
		response.InternalError(c, "failed to create brand: "+err.Error())
		return
	}

	response.Success(c, brand)
}

func (h *Handler) GetBrandByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid brand id: "+err.Error())
		return
	}

	brand, err := h.app.GetBrandByID(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get brand", "id", id, "error", err)
		response.InternalError(c, "failed to get brand: "+err.Error())
		return
	}
	if brand == nil {
		response.NotFound(c, "brand not found")
		return
	}

	response.Success(c, brand)
}

type updateBrandRequest struct {
	Name *string `json:"name"`
	Logo *string `json:"logo"`
}

func (h *Handler) UpdateBrand(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid brand id: "+err.Error())
		return
	}

	var req updateBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	brand, err := h.app.UpdateBrand(c.Request.Context(), id, req.Name, req.Logo)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to update brand", "id", id, "error", err)
		response.InternalError(c, "failed to update brand: "+err.Error())
		return
	}

	response.Success(c, brand)
}

func (h *Handler) DeleteBrand(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid brand id: "+err.Error())
		return
	}

	if err := h.app.DeleteBrand(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to delete brand", "id", id, "error", err)
		response.InternalError(c, "failed to delete brand: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) ListBrands(c *gin.Context) {
	brands, err := h.app.ListBrands(c.Request.Context())
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list brands", "error", err)
		response.InternalError(c, "failed to list brands: "+err.Error())
		return
	}

	response.Success(c, brands)
}

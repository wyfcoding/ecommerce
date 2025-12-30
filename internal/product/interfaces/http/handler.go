package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/product/application"
	"github.com/wyfcoding/pkg/response"
)

type Handler struct {
	app    *application.ProductService
	logger *slog.Logger
}

func NewHandler(app *application.ProductService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	v1 := r.Group("/product")
	{
		// Product
		v1.POST("", h.CreateProduct)
		v1.GET("/:id", h.GetProduct)
		v1.PUT("/:id", h.UpdateProduct)
		v1.DELETE("/:id", h.DeleteProduct)
		v1.GET("", h.ListProducts)

		// SKU
		v1.POST("/:id/skus", h.AddSKU)
		v1.PUT("/skus/:skuId", h.UpdateSKU)
		v1.DELETE("/skus/:skuId", h.DeleteSKU)

		// Category
		categories := v1.Group("/categories")
		{
			categories.POST("", h.CreateCategory)
			categories.GET("", h.ListCategories)
			categories.GET("/:id", h.GetCategory)
			categories.PUT("/:id", h.UpdateCategory)
			categories.DELETE("/:id", h.DeleteCategory)
		}

		// Brand
		brands := v1.Group("/brands")
		{
			brands.POST("", h.CreateBrand)
			brands.GET("", h.ListBrands)
			brands.GET("/:id", h.GetBrand)
			brands.PUT("/:id", h.UpdateBrand)
			brands.DELETE("/:id", h.DeleteBrand)
		}
	}
}

// --- Product Handlers ---

func (h *Handler) CreateProduct(c *gin.Context) {
	var req application.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	product, err := h.app.Manager.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		slog.ErrorContext(c, "create product failed", "err", err)
		response.Error(c, err)
		return
	}
	response.Success(c, product)
}

func (h *Handler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id", "")
		return
	}

	product, err := h.app.Query.GetProductByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	if product == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "product not found", "")
		return
	}
	response.Success(c, product)
}

func (h *Handler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id", "")
		return
	}

	var req application.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	product, err := h.app.Manager.UpdateProduct(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, product)
}

func (h *Handler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id", "")
		return
	}

	if err := h.app.Manager.DeleteProduct(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"status": "ok"})
}

func (h *Handler) ListProducts(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	var (
		categoryID uint64
		brandID    uint64
	)
	if val := c.Query("category_id"); val != "" {
		categoryID, err = strconv.ParseUint(val, 10, 64)
		if err != nil {
			response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid category_id", err.Error())
			return
		}
	}
	if val := c.Query("brand_id"); val != "" {
		brandID, err = strconv.ParseUint(val, 10, 64)
		if err != nil {
			response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid brand_id", err.Error())
			return
		}
	}

	products, total, err := h.app.Query.ListProducts(c.Request.Context(), page, pageSize, categoryID, brandID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithPagination(c, products, total, int32(page), int32(pageSize))
}

// --- SKU Handlers ---

func (h *Handler) AddSKU(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid product id", "")
		return
	}

	var req application.AddSKURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	sku, err := h.app.Manager.AddSKU(c.Request.Context(), productID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, sku)
}

func (h *Handler) UpdateSKU(c *gin.Context) {
	skuID, err := strconv.ParseUint(c.Param("skuId"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid sku id", "")
		return
	}

	var req application.UpdateSKURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	sku, err := h.app.Manager.UpdateSKU(c.Request.Context(), skuID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, sku)
}

func (h *Handler) DeleteSKU(c *gin.Context) {
	skuID, err := strconv.ParseUint(c.Param("skuId"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid sku id", "")
		return
	}

	if err := h.app.Manager.DeleteSKU(c.Request.Context(), skuID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"status": "ok"})
}

// --- Category Handlers ---

func (h *Handler) CreateCategory(c *gin.Context) {
	var req application.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	category, err := h.app.Manager.CreateCategory(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, category)
}

func (h *Handler) ListCategories(c *gin.Context) {
	var (
		parentID uint64
		err      error
	)
	if val := c.Query("parent_id"); val != "" {
		parentID, err = strconv.ParseUint(val, 10, 64)
		if err != nil {
			response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid parent_id", err.Error())
			return
		}
	}
	categories, err := h.app.Query.ListCategories(c.Request.Context(), parentID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, categories)
}

func (h *Handler) GetCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id", "")
		return
	}
	category, err := h.app.Query.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	if category == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "category not found", "")
		return
	}
	response.Success(c, category)
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id", "")
		return
	}
	var req application.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, err.Error(), "")
		return
	}
	category, err := h.app.Manager.UpdateCategory(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, category)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id", "")
		return
	}
	if err := h.app.Manager.DeleteCategory(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"status": "ok"})
}

// --- Brand Handlers ---

func (h *Handler) CreateBrand(c *gin.Context) {
	var req application.CreateBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	brand, err := h.app.Manager.CreateBrand(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, brand)
}

func (h *Handler) ListBrands(c *gin.Context) {
	brands, err := h.app.Query.ListBrands(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, brands)
}

func (h *Handler) GetBrand(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id", "")
		return
	}
	brand, err := h.app.Query.GetBrandByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	if brand == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "brand not found", "")
		return
	}
	response.Success(c, brand)
}

func (h *Handler) UpdateBrand(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id", "")
		return
	}
	var req application.UpdateBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, err.Error(), "")
		return
	}
	brand, err := h.app.Manager.UpdateBrand(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, brand)
}

func (h *Handler) DeleteBrand(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id", "")
		return
	}
	if err := h.app.Manager.DeleteBrand(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"status": "ok"})
}

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
	cmdService   *application.ProductCommandService
	queryService *application.ProductQueryService
	logger       *slog.Logger
}

func NewHandler(cmd *application.ProductCommandService, query *application.ProductQueryService, logger *slog.Logger) *Handler {
	return &Handler{
		cmdService:   cmd,
		queryService: query,
		logger:       logger,
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

// CreateProduct 处理商品创建请求。
func (h *Handler) CreateProduct(c *gin.Context) {
	var req application.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	product, err := h.cmdService.CreateProduct(c.Request.Context(), &application.CreateProductCommand{
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		BrandID:     req.BrandID,
		Type:        req.Type,
		Price:       req.Price,
		Stock:       req.Stock,
	})
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "create product failed", "error", err)
		response.Error(c, err)
		return
	}
	response.Success(c, product)
}

// GetProduct 获取商品详情。
func (h *Handler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid product id", "")
		return
	}

	product, err := h.queryService.GetProductByID(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "get product detail failed", "product_id", id, "error", err)
		response.Error(c, err)
		return
	}
	if product == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "product not found", "")
		return
	}
	response.Success(c, product)
}

// UpdateProduct 更新商品基本信息。
func (h *Handler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid product id", "")
		return
	}

	var req application.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid update data", err.Error())
		return
	}

	product, err := h.cmdService.UpdateProduct(c.Request.Context(), &application.UpdateProductCommand{
		ID:          uint(id),
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		BrandID:     req.BrandID,
		Status:      req.Status,
	})
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "update product failed", "product_id", id, "error", err)
		response.Error(c, err)
		return
	}
	response.Success(c, product)
}

// DeleteProduct 逻辑或物理删除商品。
func (h *Handler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid product id", "")
		return
	}

	if err := h.cmdService.DeleteProduct(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "delete product failed", "product_id", id, "error", err)
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

	products, total, err := h.queryService.ListProducts(c.Request.Context(), page, pageSize, categoryID, brandID)
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
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	sku, err := h.cmdService.AddSKU(c.Request.Context(), &application.AddSKUCommand{
		ProductID: uint(productID),
		Name:      req.Name,
		Price:     req.Price,
		Stock:     req.Stock,
		Image:     req.Image,
		Specs:     req.Specs,
	})
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "add sku failed", "product_id", productID, "error", err)
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
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	sku, err := h.cmdService.UpdateSKU(c.Request.Context(), &application.UpdateSKUCommand{
		ID:    uint(skuID),
		Price: req.Price,
		Stock: req.Stock,
		Image: req.Image,
	})
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "update sku failed", "sku_id", skuID, "error", err)
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

	if err := h.cmdService.DeleteSKU(c.Request.Context(), skuID); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "delete sku failed", "sku_id", skuID, "error", err)
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"status": "ok"})
}

// --- Category Handlers ---

func (h *Handler) CreateCategory(c *gin.Context) {
	var req application.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	category, err := h.cmdService.CreateCategory(c.Request.Context(), &application.CreateCategoryCommand{
		Name:     req.Name,
		ParentID: req.ParentID,
	})
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "create category failed", "name", req.Name, "error", err)
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
			response.ErrorWithStatus(c, http.StatusBadRequest, "invalid parent_id", err.Error())
			return
		}
	}
	categories, err := h.queryService.ListCategories(c.Request.Context(), parentID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "list categories failed", "parent_id", parentID, "error", err)
		response.Error(c, err)
		return
	}
	response.Success(c, categories)
}

func (h *Handler) GetCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid category id", "")
		return
	}
	category, err := h.queryService.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "get category failed", "id", id, "error", err)
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
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid category id", "")
		return
	}
	var req application.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}
	category, err := h.cmdService.UpdateCategory(c.Request.Context(), &application.UpdateCategoryCommand{
		ID:       uint(id),
		Name:     req.Name,
		ParentID: req.ParentID,
		Sort:     req.Sort,
	})
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "update category failed", "id", id, "error", err)
		response.Error(c, err)
		return
	}
	response.Success(c, category)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid category id", "")
		return
	}
	if err := h.cmdService.DeleteCategory(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "delete category failed", "id", id, "error", err)
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"status": "ok"})
}

// --- Brand Handlers ---

func (h *Handler) CreateBrand(c *gin.Context) {
	var req application.CreateBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	brand, err := h.cmdService.CreateBrand(c.Request.Context(), &application.CreateBrandCommand{
		Name: req.Name,
		Logo: req.Logo,
	})
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "create brand failed", "name", req.Name, "error", err)
		response.Error(c, err)
		return
	}
	response.Success(c, brand)
}

func (h *Handler) ListBrands(c *gin.Context) {
	brands, err := h.queryService.ListBrands(c.Request.Context())
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "list brands failed", "error", err)
		response.Error(c, err)
		return
	}
	response.Success(c, brands)
}

func (h *Handler) GetBrand(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid brand id", "")
		return
	}
	brand, err := h.queryService.GetBrandByID(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "get brand failed", "id", id, "error", err)
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
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid brand id", "")
		return
	}
	var req application.UpdateBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}
	brand, err := h.cmdService.UpdateBrand(c.Request.Context(), &application.UpdateBrandCommand{
		ID:   uint(id),
		Name: req.Name,
		Logo: req.Logo,
	})
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "update brand failed", "id", id, "error", err)
		response.Error(c, err)
		return
	}
	response.Success(c, brand)
}

func (h *Handler) DeleteBrand(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid brand id", "")
		return
	}
	if err := h.cmdService.DeleteBrand(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "delete brand failed", "id", id, "error", err)
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"status": "ok"})
}

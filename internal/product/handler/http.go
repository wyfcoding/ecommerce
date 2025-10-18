package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/product/model"
	"ecommerce/internal/product/service"
)

// ProductHandler 负责处理商品的 HTTP 请求
type ProductHandler struct {
	svc    service.ProductService
	logger *zap.Logger
}

// NewProductHandler 创建一个新的 ProductHandler 实例
func NewProductHandler(svc service.ProductService, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{svc: svc, logger: logger}
}

// RegisterRoutes 在 Gin 引擎上注册所有商品相关的路由
func (h *ProductHandler) RegisterRoutes(r *gin.Engine) {
	group := r.Group("/api/v1/products")
	{
		group.POST("", h.CreateProduct)        // 创建商品
		group.GET("/:id", h.GetProduct)         // 获取单个商品
		group.GET("", h.ListProducts)          // 列出商品 (分页)
		group.PUT("/:id", h.UpdateProduct)       // 更新商品
		group.DELETE("/:id", h.DeleteProduct)    // 删除商品
		group.GET("/search", h.SearchProducts) // 搜索商品
	}
}

// CreateProduct 处理创建商品的请求
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var product model.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		h.logger.Error("Invalid request body for CreateProduct", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	createdProduct, err := h.svc.CreateProduct(c.Request.Context(), &product)
	if err != nil {
		h.logger.Error("Failed to create product", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建商品失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "商品创建成功", "product": createdProduct})
}

// GetProduct 处理获取单个商品的请求
func (h *ProductHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的商品ID"})
		return
	}

	product, err := h.svc.GetProduct(c.Request.Context(), uint(id))
	if err != nil {
		// 根据错误类型返回不同的状态码
		if err.Error() == "商品不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取商品失败: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, product)
}

// ListProducts 处理列出商品的请求 (支持分页和过滤)
func (h *ProductHandler) ListProducts(c *gin.Context) {
	// 从查询参数获取分页信息
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// 简单的过滤参数
	var categoryID, brandID *uint
	if catIDStr := c.Query("category_id"); catIDStr != "" {
		catID, err := strconv.ParseUint(catIDStr, 10, 32)
		if err == nil {
			cid := uint(catID)
			categoryID = &cid
		}
	}

	products, total, err := h.svc.ListProducts(c.Request.Context(), page, pageSize, categoryID, brandID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取商品列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":       page,
		"pageSize":   pageSize,
		"total":      total,
		"products":   products,
	})
}

// UpdateProduct ... (待实现)
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "功能待实现"})
}

// DeleteProduct 处理删除商品的请求
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的商品ID"})
		return
	}

	if err := h.svc.DeleteProduct(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除商品失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "商品删除成功"})
}

// SearchProducts ... (待实现)
func (h *ProductHandler) SearchProducts(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "功能待实现"})
}

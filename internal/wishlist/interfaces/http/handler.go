package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字轉換工具。

	"github.com/wyfcoding/ecommerce/internal/wishlist/application" // 导入收藏夹模块的应用服务。
	"github.com/wyfcoding/pkg/response"                            // 导入统一的响应处理工具。

	"log/slog" // 导入结构化日志库。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
)

// Handler 结构体定义了Wishlist模块的HTTP处理层。
type Handler struct {
	app    *application.Wishlist // 依赖Wishlist应用服务 (Facade)。
	logger *slog.Logger          // 日志记录器。
}

// NewHandler 创建并返回一个新的 Wishlist HTTP Handler 实例。
func NewHandler(app *application.Wishlist, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// Add 处理将商品添加到收藏夹的HTTP请求。
func (h *Handler) Add(c *gin.Context) {
	var req struct {
		UserID      uint64 `json:"user_id" binding:"required"`
		ProductID   uint64 `json:"product_id" binding:"required"`
		SkuID       uint64 `json:"sku_id" binding:"required"`
		ProductName string `json:"product_name" binding:"required"`
		SkuName     string `json:"sku_name" binding:"required"`
		Price       uint64 `json:"price" binding:"required"`
		ImageURL    string `json:"image_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	wishlist, err := h.app.Add(c.Request.Context(), req.UserID, req.ProductID, req.SkuID, req.ProductName, req.SkuName, req.Price, req.ImageURL)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to add to wishlist", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add to wishlist", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Added to wishlist successfully", wishlist)
}

// Remove 处理从收藏夹移除商品的HTTP请求。
func (h *Handler) Remove(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)

	if err := h.app.Remove(c.Request.Context(), userID, id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to remove from wishlist", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to remove from wishlist", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Removed from wishlist successfully", nil)
}

// List 处理获取用户收藏夹列表的HTTP请求。
func (h *Handler) List(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.app.List(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list wishlist", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list wishlist", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Wishlist listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CheckStatus 处理检查商品是否在用户收藏夹中的HTTP请求。
func (h *Handler) CheckStatus(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	skuID, _ := strconv.ParseUint(c.Query("sku_id"), 10, 64)

	exists, err := h.app.CheckStatus(c.Request.Context(), userID, skuID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to check status", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to check status", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Status checked successfully", gin.H{"exists": exists})
}

// Clear 处理清空收藏夹的HTTP请求。
func (h *Handler) Clear(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)

	if err := h.app.Clear(c.Request.Context(), userID); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to clear wishlist", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to clear wishlist", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Wishlist cleared successfully", nil)
}

// RemoveByProduct 处理按商品移除收藏夹条目的HTTP请求。
func (h *Handler) RemoveByProduct(c *gin.Context) {
	skuID, _ := strconv.ParseUint(c.Param("sku_id"), 10, 64)
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)

	if err := h.app.RemoveByProduct(c.Request.Context(), userID, skuID); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to remove from wishlist by product", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to remove from wishlist by product", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Removed from wishlist by product successfully", nil)
}

// RegisterRoutes 注册路由.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/wishlist")
	{
		group.POST("", h.Add)
		group.DELETE("/:id", h.Remove)
		group.GET("", h.List)
		group.GET("/status", h.CheckStatus)
		group.DELETE("", h.Clear)
		group.DELETE("/product/:sku_id", h.RemoveByProduct)
	}
}

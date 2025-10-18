package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/wishlist/service"
	// 伪代码: 模拟认证中间件
	// auth "ecommerce/internal/auth/handler"
)

// WishlistHandler 负责处理心愿单的 HTTP 请求
type WishlistHandler struct {
	svc    service.WishlistService
	logger *zap.Logger
}

// NewWishlistHandler 创建一个新的 WishlistHandler 实例
func NewWishlistHandler(svc service.WishlistService, logger *zap.Logger) *WishlistHandler {
	return &WishlistHandler{svc: svc, logger: logger}
}

// RegisterRoutes 在 Gin 引擎上注册所有心愿单相关的路由
func (h *WishlistHandler) RegisterRoutes(r *gin.Engine) {
	// 所有端点都需要用户认证
	group := r.Group("/api/v1/wishlist")
	// group.Use(auth.AuthMiddleware(...))
	{
		group.GET("", h.ListItems)
		group.POST("", h.AddItem)
		group.DELETE("/:product_id", h.RemoveItem)
	}
}

// ListItems 处理获取用户心愿单列表的请求
func (h *WishlistHandler) ListItems(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	items, err := h.svc.ListItems(c.Request.Context(), userID.(uint))
	if err != nil {
		h.logger.Error("Failed to list wishlist items", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取心愿单列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"wishlist": items})
}

// AddItemRequest 定义了添加心愿单项目的请求体
type AddItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
}

// AddItem 处理用户添加心愿单项目的请求
func (h *WishlistHandler) AddItem(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	var req AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	item, err := h.svc.AddItem(c.Request.Context(), userID.(uint), req.ProductID)
	if err != nil {
		h.logger.Error("Failed to add item to wishlist", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "已添加到心愿单", "item": item})
}

// RemoveItem 处理用户移除心愿单项目的请求
func (h *WishlistHandler) RemoveItem(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的商品ID"})
		return
	}

	if err := h.svc.RemoveItem(c.Request.Context(), userID.(uint), uint(productID)); err != nil {
		h.logger.Error("Failed to remove item from wishlist", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "移除失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已从心愿单移除"})
}
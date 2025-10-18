package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/cart/service"
)

// CartHandler 负责处理购物车的 HTTP 请求
type CartHandler struct {
	svc    service.CartService
	logger *zap.Logger
}

// NewCartHandler 创建一个新的 CartHandler 实例
func NewCartHandler(svc service.CartService, logger *zap.Logger) *CartHandler {
	return &CartHandler{svc: svc, logger: logger}
}

// RegisterRoutes 在 Gin 引擎上注册所有购物车相关的路由
func (h *CartHandler) RegisterRoutes(r *gin.Engine) {
	// 创建一个 /api/v1/cart 路由组
	cartGroup := r.Group("/api/v1/cart")
	{
		// GET /: 获取当前用户的购物车
		cartGroup.GET("", h.GetCart)
		// POST /items: 向购物车添加商品
		cartGroup.POST("/items", h.AddItem)
		// PUT /items/:product_id: 更新购物车中商品的数量
		cartGroup.PUT("/items/:product_id", h.UpdateItemQuantity)
		// DELETE /items/:product_id: 从购物车删除商品
		cartGroup.DELETE("/items/:product_id", h.RemoveItem)
		// DELETE /: 清空购物车
		cartGroup.DELETE("", h.ClearCart)
	}
}

// AddItemRequest 定义了添加商品到购物车的请求体结构
type AddItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

// AddItem 处理添加商品的请求
// @Summary 添加商品到购物车
// @Description 将指定数量的商品添加到当前用户的购物车中
// @Tags Cart
// @Accept json
// @Produce json
// @Param body body AddItemRequest true "请求体"
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 400 {object} map[string]string "请求错误"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/cart/items [post]
func (h *CartHandler) AddItem(c *gin.Context) {
	// 从 Gin 上下文中获取 userID，这里假设有中间件已经完成了用户认证并注入了 userID
	// 在实际应用中，这通常从 JWT token 中解析得到
	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Warn("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	var req AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body for AddItem", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	cart, err := h.svc.AddItem(c.Request.Context(), userID.(uint), req.ProductID, req.Quantity)
	if err != nil {
		h.logger.Error("Failed to add item to cart", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加商品失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "商品添加成功", "cart": cart})
}

// GetCart 处理获取购物车内容的请求
// @Summary 获取购物车
// @Description 获取当前用户的购物车所有内容及总价
// @Tags Cart
// @Produce json
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 401 {object} map[string]string "未认证"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/cart [get]
func (h *CartHandler) GetCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Warn("User ID not found in context for GetCart")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	cart, totalPrice, err := h.svc.GetCart(c.Request.Context(), userID.(uint))
	if err != nil {
		h.logger.Error("Failed to get cart", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取购物车失败: " + err.Error()})
		return
	}

	if cart == nil {
		c.JSON(http.StatusOK, gin.H{"message": "购物车是空的", "total_price": 0, "items": []string{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cart": cart, "total_price": totalPrice})
}

// UpdateItemQuantity ... (待实现)
func (h *CartHandler) UpdateItemQuantity(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "功能待实现"})
}

// RemoveItem ... (待实现)
func (h *CartHandler) RemoveItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "功能待实现"})
}

// ClearCart ... (待实现)
func (h *CartHandler) ClearCart(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "功能待实现"})
}

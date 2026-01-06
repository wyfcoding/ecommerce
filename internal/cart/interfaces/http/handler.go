package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/cart/application"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	app    *application.CartService
	logger *slog.Logger
}

func NewHandler(app *application.CartService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// GetCart 获取指定用户的购物车详情。
func (h *Handler) GetCart(c *gin.Context) {
	userIDStr := c.Query("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || userID == 0 {
		response.ErrorWithStatus(c, http.StatusBadRequest, "valid user_id query parameter is required", "")
		return
	}

	cart, err := h.app.GetCart(c.Request.Context(), userID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get cart", "user_id", userID, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, cart)
}

// AddItem 将指定商品及 SKU 添加到用户的购物车。
func (h *Handler) AddItem(c *gin.Context) {
	// ... (请求参数定义保持不变) ...
	var req struct {
		UserID          uint64  `json:"user_id" binding:"required"`
		ProductID       uint64  `json:"product_id" binding:"required"`
		SkuID           uint64  `json:"sku_id" binding:"required"`
		ProductName     string  `json:"product_name" binding:"required"`
		SkuName         string  `json:"sku_name" binding:"required"`
		Price           float64 `json:"price" binding:"required"`
		Quantity        int32   `json:"quantity" binding:"required"`
		ProductImageURL string  `json:"product_image_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	err := h.app.AddItem(c.Request.Context(), req.UserID, req.ProductID, req.SkuID, req.ProductName, req.SkuName, req.Price, req.Quantity, req.ProductImageURL)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to add item to cart", "user_id", req.UserID, "sku_id", req.SkuID, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// UpdateItemQuantity 修改购物车中特定 SKU 的购买数量。
func (h *Handler) UpdateItemQuantity(c *gin.Context) {
	var req struct {
		UserID   uint64 `json:"user_id" binding:"required"`
		SkuID    uint64 `json:"sku_id" binding:"required"`
		Quantity int32  `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	err := h.app.UpdateItemQuantity(c.Request.Context(), req.UserID, req.SkuID, req.Quantity)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to update item quantity", "user_id", req.UserID, "sku_id", req.SkuID, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// RemoveItem 从购物车中移除指定的商品项。
func (h *Handler) RemoveItem(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
		SkuID  uint64 `json:"sku_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	err := h.app.RemoveItem(c.Request.Context(), req.UserID, req.SkuID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to remove item from cart", "user_id", req.UserID, "sku_id", req.SkuID, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// ClearCart 清空指定用户的整个购物车。
func (h *Handler) ClearCart(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	err := h.app.ClearCart(c.Request.Context(), req.UserID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to clear cart", "user_id", req.UserID, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/cart")
	{
		group.GET("", h.GetCart)
		group.POST("/items", h.AddItem)
		group.PUT("/items", h.UpdateItemQuantity)
		group.DELETE("/items", h.RemoveItem)
		group.DELETE("", h.ClearCart)
	}
}

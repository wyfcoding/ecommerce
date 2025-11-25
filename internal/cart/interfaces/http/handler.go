package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/cart/application"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.CartService
	logger  *slog.Logger
}

func NewHandler(service *application.CartService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetCart 获取购物车
func (h *Handler) GetCart(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if userID == 0 {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid user ID", "user_id is required")
		return
	}

	cart, err := h.service.GetCart(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get cart", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get cart", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Cart retrieved successfully", cart)
}

// AddItem 添加商品到购物车
func (h *Handler) AddItem(c *gin.Context) {
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
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	err := h.service.AddItem(c.Request.Context(), req.UserID, req.ProductID, req.SkuID, req.ProductName, req.SkuName, req.Price, req.Quantity, req.ProductImageURL)
	if err != nil {
		h.logger.Error("Failed to add item to cart", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add item to cart", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Item added to cart successfully", nil)
}

// UpdateItemQuantity 更新购物车项数量
func (h *Handler) UpdateItemQuantity(c *gin.Context) {
	var req struct {
		UserID   uint64 `json:"user_id" binding:"required"`
		SkuID    uint64 `json:"sku_id" binding:"required"`
		Quantity int32  `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	err := h.service.UpdateItemQuantity(c.Request.Context(), req.UserID, req.SkuID, req.Quantity)
	if err != nil {
		h.logger.Error("Failed to update item quantity", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update item quantity", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Item quantity updated successfully", nil)
}

// RemoveItem 移除购物车项
func (h *Handler) RemoveItem(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
		SkuID  uint64 `json:"sku_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	err := h.service.RemoveItem(c.Request.Context(), req.UserID, req.SkuID)
	if err != nil {
		h.logger.Error("Failed to remove item from cart", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to remove item from cart", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Item removed from cart successfully", nil)
}

// ClearCart 清空购物车
func (h *Handler) ClearCart(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	err := h.service.ClearCart(c.Request.Context(), req.UserID)
	if err != nil {
		h.logger.Error("Failed to clear cart", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to clear cart", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Cart cleared successfully", nil)
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

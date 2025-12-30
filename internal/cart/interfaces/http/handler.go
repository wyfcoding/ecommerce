package http

import (
	"log/slog"
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

func (h *Handler) GetCart(c *gin.Context) {
	userIDStr := c.Query("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || userID == 0 {
		response.BadRequest(c, "valid user_id is required")
		return
	}

	cart, err := h.app.GetCart(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get cart", "user_id", userID, "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, cart)
}

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
		response.BadRequest(c, "invalid input: "+err.Error())
		return
	}

	err := h.app.AddItem(c.Request.Context(), req.UserID, req.ProductID, req.SkuID, req.ProductName, req.SkuName, req.Price, req.Quantity, req.ProductImageURL)
	if err != nil {
		h.logger.Error("Failed to add item to cart", "user_id", req.UserID, "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) UpdateItemQuantity(c *gin.Context) {
	var req struct {
		UserID   uint64 `json:"user_id" binding:"required"`
		SkuID    uint64 `json:"sku_id" binding:"required"`
		Quantity int32  `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err := h.app.UpdateItemQuantity(c.Request.Context(), req.UserID, req.SkuID, req.Quantity)
	if err != nil {
		h.logger.Error("Failed to update quantity", "user_id", req.UserID, "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) RemoveItem(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
		SkuID  uint64 `json:"sku_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err := h.app.RemoveItem(c.Request.Context(), req.UserID, req.SkuID)
	if err != nil {
		h.logger.Error("Failed to remove item", "user_id", req.UserID, "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) ClearCart(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err := h.app.ClearCart(c.Request.Context(), req.UserID)
	if err != nil {
		h.logger.Error("Failed to clear cart", "user_id", req.UserID, "error", err)
		response.InternalError(c, err.Error())
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

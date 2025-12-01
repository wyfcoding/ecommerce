package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/wishlist/application"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"log/slog"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *application.WishlistService
	logger  *slog.Logger
}

func NewHandler(service *application.WishlistService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Add 添加收藏
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

	wishlist, err := h.service.Add(c.Request.Context(), req.UserID, req.ProductID, req.SkuID, req.ProductName, req.SkuName, req.Price, req.ImageURL)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to add to wishlist", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add to wishlist", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Added to wishlist successfully", wishlist)
}

// Remove 移除收藏
func (h *Handler) Remove(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// Ideally userID should come from context/token, but for now taking from query or assuming passed
	// In a real scenario, we'd extract userID from the auth token.
	// Here we'll assume it's passed as a query param for simplicity in this refactor step,
	// or we can just delete by ID if we trust the ID.
	// Let's require user_id in query for safety.
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "User ID required", err.Error())
		return
	}

	if err := h.service.Remove(c.Request.Context(), userID, id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to remove from wishlist", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to remove from wishlist", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Removed from wishlist successfully", nil)
}

// List 收藏列表
func (h *Handler) List(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "User ID required", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.List(c.Request.Context(), userID, page, pageSize)
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

// CheckStatus 检查状态
func (h *Handler) CheckStatus(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "User ID required", err.Error())
		return
	}
	skuID, err := strconv.ParseUint(c.Query("sku_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "SKU ID required", err.Error())
		return
	}

	exists, err := h.service.CheckStatus(c.Request.Context(), userID, skuID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to check status", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to check status", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Status checked successfully", gin.H{"exists": exists})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/wishlist")
	{
		group.POST("", h.Add)
		group.DELETE("/:id", h.Remove)
		group.GET("", h.List)
		group.GET("/status", h.CheckStatus)
	}
}

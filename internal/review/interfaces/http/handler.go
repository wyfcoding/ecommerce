package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/review/application"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.ReviewService
	logger  *slog.Logger
}

func NewHandler(service *application.ReviewService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateReview 创建评论
func (h *Handler) CreateReview(c *gin.Context) {
	var req struct {
		UserID    uint64   `json:"user_id" binding:"required"`
		ProductID uint64   `json:"product_id" binding:"required"`
		OrderID   uint64   `json:"order_id" binding:"required"`
		SkuID     uint64   `json:"sku_id" binding:"required"`
		Rating    int      `json:"rating" binding:"required,min=1,max=5"`
		Content   string   `json:"content" binding:"required"`
		Images    []string `json:"images"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	review, err := h.service.CreateReview(c.Request.Context(), req.UserID, req.ProductID, req.OrderID, req.SkuID, req.Rating, req.Content, req.Images)
	if err != nil {
		h.logger.Error("Failed to create review", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create review", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Review created successfully", review)
}

// ApproveReview 审核通过
func (h *Handler) ApproveReview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Review ID", err.Error())
		return
	}

	if err := h.service.ApproveReview(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to approve review", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to approve review", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Review approved successfully", nil)
}

// RejectReview 审核拒绝
func (h *Handler) RejectReview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Review ID", err.Error())
		return
	}

	if err := h.service.RejectReview(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to reject review", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to reject review", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Review rejected successfully", nil)
}

// ListReviews 评论列表
func (h *Handler) ListReviews(c *gin.Context) {
	productID, _ := strconv.ParseUint(c.Query("product_id"), 10, 64)
	statusStr := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var status *int
	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil {
			status = &s
		}
	}

	list, total, err := h.service.ListReviews(c.Request.Context(), productID, status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list reviews", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list reviews", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Reviews listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetProductStats 获取商品评分统计
func (h *Handler) GetProductStats(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("product_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Product ID", err.Error())
		return
	}

	stats, err := h.service.GetProductStats(c.Request.Context(), productID)
	if err != nil {
		h.logger.Error("Failed to get product stats", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get product stats", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Product stats retrieved successfully", stats)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/reviews")
	{
		group.POST("", h.CreateReview)
		group.GET("", h.ListReviews)
		group.PUT("/:id/approve", h.ApproveReview)
		group.PUT("/:id/reject", h.RejectReview)
		group.GET("/stats/:product_id", h.GetProductStats)
	}
}

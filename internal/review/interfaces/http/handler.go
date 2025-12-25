package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字轉換工具。

	"github.com/wyfcoding/ecommerce/internal/review/application" // 导入评论模块的应用服务。
	"github.com/wyfcoding/pkg/response"                          // 导入统一的响应处理工具。

	"log/slog" // 导入结构化日志库。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
)

// Handler 结构体定义了Review模块的HTTP处理层。
type Handler struct {
	app    *application.Review // 依赖Review应用服务 (Facade)。
	logger *slog.Logger               // 日志记录器。
}

// NewHandler 创建并返回一个新的 Review HTTP Handler 实例。
func NewHandler(app *application.Review, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// CreateReview 处理创建评论的HTTP请求。
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

	review, err := h.app.CreateReview(c.Request.Context(), req.UserID, req.ProductID, req.OrderID, req.SkuID, req.Rating, req.Content, req.Images)
	if err != nil {
		h.logger.Error("Failed to create review", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create review", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Review created successfully", review)
}

// AuditReview 处理审核评论的HTTP请求.
func (h *Handler) AuditReview(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Approved bool `json:"approved"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.app.AuditReview(c.Request.Context(), id, req.Approved); err != nil {
		h.logger.Error("Failed to audit review", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to audit review", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Review audited successfully", nil)
}

// ListReviews 获取评论列表.
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

	list, total, err := h.app.ListReviews(c.Request.Context(), productID, status, page, pageSize)
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

// GetProductStats 获取商品评分统计.
func (h *Handler) GetProductStats(c *gin.Context) {
	productID, _ := strconv.ParseUint(c.Param("product_id"), 10, 64)
	stats, err := h.app.GetProductStats(c.Request.Context(), productID)
	if err != nil {
		h.logger.Error("Failed to get product stats", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get product stats", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Product stats retrieved successfully", stats)
}

// RegisterRoutes 注册路由.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/reviews")
	{
		group.POST("", h.CreateReview)
		group.GET("", h.ListReviews)
		group.PUT("/:id/audit", h.AuditReview)
		group.GET("/stats/:product_id", h.GetProductStats)
	}
}

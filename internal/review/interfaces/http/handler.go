package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/review/application" // 导入评论模块的应用服务。
	"github.com/wyfcoding/ecommerce/pkg/response"                // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了Review模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.ReviewService // 依赖Review应用服务，处理核心业务逻辑。
	logger  *slog.Logger               // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Review HTTP Handler 实例。
func NewHandler(service *application.ReviewService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateReview 处理创建评论的HTTP请求。
// Method: POST
// Path: /reviews
func (h *Handler) CreateReview(c *gin.Context) {
	// 定义请求体结构，用于接收评论的创建信息。
	var req struct {
		UserID    uint64   `json:"user_id" binding:"required"`            // 用户ID，必填。
		ProductID uint64   `json:"product_id" binding:"required"`         // 商品ID，必填。
		OrderID   uint64   `json:"order_id" binding:"required"`           // 订单ID，必填。
		SkuID     uint64   `json:"sku_id" binding:"required"`             // SKU ID，必填。
		Rating    int      `json:"rating" binding:"required,min=1,max=5"` // 评分，必填，范围1-5。
		Content   string   `json:"content" binding:"required"`            // 评论内容，必填。
		Images    []string `json:"images"`                                // 图片URL列表，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建评论。
	review, err := h.service.CreateReview(c.Request.Context(), req.UserID, req.ProductID, req.OrderID, req.SkuID, req.Rating, req.Content, req.Images)
	if err != nil {
		h.logger.Error("Failed to create review", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create review", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Review created successfully", review)
}

// ApproveReview 处理审核通过评论的HTTP请求。
// Method: PUT
// Path: /reviews/:id/approve
func (h *Handler) ApproveReview(c *gin.Context) {
	// 从URL路径中解析评论ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Review ID", err.Error())
		return
	}

	// 调用应用服务层批准评论。
	if err := h.service.ApproveReview(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to approve review", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to approve review", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Review approved successfully", nil)
}

// RejectReview 处理审核拒绝评论的HTTP请求。
// Method: PUT
// Path: /reviews/:id/reject
func (h *Handler) RejectReview(c *gin.Context) {
	// 从URL路径中解析评论ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Review ID", err.Error())
		return
	}

	// 调用应用服务层拒绝评论。
	if err := h.service.RejectReview(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to reject review", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to reject review", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Review rejected successfully", nil)
}

// ListReviews 处理获取评论列表的HTTP请求。
// Method: GET
// Path: /reviews
func (h *Handler) ListReviews(c *gin.Context) {
	// 从查询参数中获取商品ID、状态、页码和每页大小，并设置默认值。
	productID, _ := strconv.ParseUint(c.Query("product_id"), 10, 64)
	statusStr := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var status *int
	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil { // 只有当状态字符串能成功转换为int时才设置过滤状态。
			status = &s
		}
	}

	// 调用应用服务层获取评论列表。
	list, total, err := h.service.ListReviews(c.Request.Context(), productID, status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list reviews", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list reviews", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Reviews listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetProductStats 处理获取商品评分统计的HTTP请求。
// Method: GET
// Path: /reviews/stats/:product_id
func (h *Handler) GetProductStats(c *gin.Context) {
	// 从URL路径中解析商品ID。
	productID, err := strconv.ParseUint(c.Param("product_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Product ID", err.Error())
		return
	}

	// 调用应用服务层获取商品评分统计。
	stats, err := h.service.GetProductStats(c.Request.Context(), productID)
	if err != nil {
		h.logger.Error("Failed to get product stats", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get product stats", err.Error())
		return
	}

	// 返回成功的响应，包含商品评分统计数据。
	response.SuccessWithStatus(c, http.StatusOK, "Product stats retrieved successfully", stats)
}

// RegisterRoutes 在给定的Gin路由组中注册Review模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /reviews 路由组，用于所有评论相关接口。
	group := r.Group("/reviews")
	{
		group.POST("", h.CreateReview)                     // 创建评论。
		group.GET("", h.ListReviews)                       // 获取评论列表。
		group.PUT("/:id/approve", h.ApproveReview)         // 审核通过评论。
		group.PUT("/:id/reject", h.RejectReview)           // 审核拒绝评论。
		group.GET("/stats/:product_id", h.GetProductStats) // 获取商品评分统计。
		// TODO: 补充获取评论详情、删除评论、用户评论列表等接口。
	}
}

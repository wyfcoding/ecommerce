package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/review/service"
	// 伪代码: 模拟认证中间件
	// auth "ecommerce/internal/auth/handler"
)

// ReviewHandler 负责处理评论的 HTTP 请求
type ReviewHandler struct {
	svc    service.ReviewService
	logger *zap.Logger
}

// NewReviewHandler 创建一个新的 ReviewHandler 实例
func NewReviewHandler(svc service.ReviewService, logger *zap.Logger) *ReviewHandler {
	return &ReviewHandler{svc: svc, logger: logger}
}

// RegisterRoutes 在 Gin 引擎上注册所有评论相关的路由
func (h *ReviewHandler) RegisterRoutes(r *gin.Engine) {
	// 公开的路由
	pubGroup := r.Group("/api/v1")
	{
		pubGroup.GET("/products/:product_id/reviews", h.ListReviews)
	}

	// 需要用户认证的路由
	authGroup := r.Group("/api/v1/reviews")
	// authGroup.Use(auth.AuthMiddleware(...))
	{
		authGroup.POST("", h.CreateReview)
		authGroup.POST("/:review_id/comments", h.AddComment)
	}

	// 需要管理员权限的路由
	adminGroup := r.Group("/api/v1/admin/reviews")
	// adminGroup.Use(auth.AuthMiddleware(...), auth.AdminMiddleware(...))
	{
		adminGroup.PUT("/:review_id/approve", h.ApproveReview)
	}
}

// CreateReviewRequest 定义了创建评论的请求体
type CreateReviewRequest struct {
	ProductID uint   `json:"product_id" binding:"required"`
	OrderID   uint   `json:"order_id" binding:"required"`
	Rating    int    `json:"rating" binding:"required,gte=1,lte=5"`
	Title     string `json:"title"`
	Content   string `json:"content" binding:"required"`
}

// CreateReview 处理用户提交评论的请求
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID, _ := c.Get("userID") // 假设中间件已注入 userID

	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	review, err := h.svc.CreateReview(c.Request.Context(), userID.(uint), req.ProductID, req.OrderID, req.Rating, req.Title, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建评论失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "评论已提交，等待审核", "review": review})
}

// ListReviews 处理获取商品评论列表的请求
func (h *ReviewHandler) ListReviews(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的商品ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	reviews, total, err := h.svc.ListProductReviews(c.Request.Context(), uint(productID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取评论列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":     page,
		"pageSize": pageSize,
		"total":    total,
		"reviews":  reviews,
	})
}

// AddComment ... (待实现)
func (h *ReviewHandler) AddComment(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "功能待实现"})
}

// ApproveReview 处理管理员审核通过评论的请求
func (h *ReviewHandler) ApproveReview(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的评论ID"})
		return
	}

	if err := h.svc.ApproveReview(c.Request.Context(), uint(reviewID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审核评论失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "评论已审核通过"})
}
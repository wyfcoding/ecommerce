package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/coupon/model"
	"ecommerce/internal/coupon/service"
	// 伪代码: 模拟认证中间件
	// auth "ecommerce/internal/auth/handler"
)

// CouponHandler 负责处理优惠券的 HTTP 请求
type CouponHandler struct {
	svc    service.CouponService
	logger *zap.Logger
}

// NewCouponHandler 创建一个新的 CouponHandler 实例
func NewCouponHandler(svc service.CouponService, logger *zap.Logger) *CouponHandler {
	return &CouponHandler{svc: svc, logger: logger}
}

// RegisterRoutes 在 Gin 引擎上注册所有优惠券相关的路由
func (h *CouponHandler) RegisterRoutes(r *gin.Engine) {
	// 所有端点都需要用户认证
	group := r.Group("/api/v1/coupons")
	// group.Use(auth.AuthMiddleware(...))
	{
		group.GET("/my-coupons", h.ListMyCoupons) // 获取我的优惠券列表
		group.POST("/claim", h.ClaimCoupon)     // 用户领取优惠券
	}
}

// ListMyCoupons 处理获取用户优惠券列表的请求
func (h *CouponHandler) ListMyCoupons(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 从查询参数获取状态过滤器
	status := c.DefaultQuery("status", "UNUSED") // 默认只看未使用的

	coupons, err := h.svc.GetUserCoupons(c.Request.Context(), userID.(uint), model.CouponStatus(status))
	if err != nil {
		h.logger.Error("Failed to get user coupons", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取优惠券列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"coupons": coupons})
}

// ClaimCouponRequest 定义了用户领取优惠券的请求体
type ClaimCouponRequest struct {
	CouponCode string `json:"coupon_code" binding:"required"`
}

// ClaimCoupon 处理用户领取优惠券的请求
func (h *CouponHandler) ClaimCoupon(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	var req ClaimCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	userCoupon, err := h.svc.AssignCouponToUser(c.Request.Context(), userID.(uint), req.CouponCode)
	if err != nil {
		h.logger.Error("Failed to assign coupon to user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "领取优惠券失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "领取成功", "coupon": userCoupon})
}
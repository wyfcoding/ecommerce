package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/coupon/application"
	"github.com/wyfcoding/ecommerce/internal/coupon/domain/entity"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.CouponService
	logger  *slog.Logger
}

func NewHandler(service *application.CouponService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateCoupon 创建优惠券
func (h *Handler) CreateCoupon(c *gin.Context) {
	var req struct {
		Name           string            `json:"name" binding:"required"`
		Description    string            `json:"description"`
		Type           entity.CouponType `json:"type" binding:"required"`
		DiscountAmount int64             `json:"discount_amount" binding:"required"`
		MinOrderAmount int64             `json:"min_order_amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	coupon, err := h.service.CreateCoupon(c.Request.Context(), req.Name, req.Description, req.Type, req.DiscountAmount, req.MinOrderAmount)
	if err != nil {
		h.logger.Error("Failed to create coupon", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create coupon", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Coupon created successfully", coupon)
}

// ActivateCoupon 激活优惠券
func (h *Handler) ActivateCoupon(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	err = h.service.ActivateCoupon(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to activate coupon", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to activate coupon", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Coupon activated successfully", nil)
}

// IssueCoupon 发放优惠券
func (h *Handler) IssueCoupon(c *gin.Context) {
	var req struct {
		UserID   uint64 `json:"user_id" binding:"required"`
		CouponID uint64 `json:"coupon_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	userCoupon, err := h.service.IssueCoupon(c.Request.Context(), req.UserID, req.CouponID)
	if err != nil {
		h.logger.Error("Failed to issue coupon", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to issue coupon", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Coupon issued successfully", userCoupon)
}

// ListCoupons 获取优惠券列表
func (h *Handler) ListCoupons(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListCoupons(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list coupons", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list coupons", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Coupons listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListUserCoupons 获取用户优惠券列表
func (h *Handler) ListUserCoupons(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if userID == 0 {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid user ID", "user_id is required")
		return
	}
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListUserCoupons(c.Request.Context(), userID, status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list user coupons", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list user coupons", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "User coupons listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateActivity 创建活动
func (h *Handler) CreateActivity(c *gin.Context) {
	var req struct {
		Name        string    `json:"name" binding:"required"`
		Description string    `json:"description"`
		StartTime   time.Time `json:"start_time" binding:"required"`
		EndTime     time.Time `json:"end_time" binding:"required"`
		CouponIDs   []uint64  `json:"coupon_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	activity, err := h.service.CreateActivity(c.Request.Context(), req.Name, req.Description, req.StartTime, req.EndTime, req.CouponIDs)
	if err != nil {
		h.logger.Error("Failed to create activity", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create activity", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Activity created successfully", activity)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/coupons")
	{
		group.POST("", h.CreateCoupon)
		group.PUT("/:id/activate", h.ActivateCoupon)
		group.GET("", h.ListCoupons)
		group.POST("/issue", h.IssueCoupon)
	}

	userGroup := r.Group("/user-coupons")
	{
		userGroup.GET("", h.ListUserCoupons)
	}

	activityGroup := r.Group("/coupon-activities")
	{
		activityGroup.POST("", h.CreateActivity)
	}
}

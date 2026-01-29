package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。
	"time"     // 导入时间包，用于时间解析。

	"github.com/wyfcoding/ecommerce/internal/coupon/application" // 导入优惠券模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/coupon/domain"      // 导入优惠券模块的领域层。
	"github.com/wyfcoding/pkg/response"                          // 导入统一的响应处理工具。

	"log/slog" // 导入结构化日志库。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
)

// Handler 结构体定义了Coupon模块的HTTP处理层。
type Handler struct {
	app    *application.Coupon // 依赖Coupon应用服务（Facade）。
	logger *slog.Logger        // 日志记录器。
}

// NewHandler 创建并返回一个新的 Coupon HTTP Handler 实例。
func NewHandler(app *application.Coupon, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// CreateCoupon 处理创建优惠券模板的 HTTP 请求。
func (h *Handler) CreateCoupon(c *gin.Context) {
	var req struct {
		Name           string `json:"name" binding:"required"`
		Description    string `json:"description"`
		Type           int    `json:"type" binding:"required"`
		DiscountAmount int64  `json:"discount_amount" binding:"required"`
		MinOrderAmount int64  `json:"min_order_amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	coupon, err := h.app.CreateCoupon(c.Request.Context(), req.Name, req.Description, req.Type, req.DiscountAmount, req.MinOrderAmount)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to create coupon", "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "coupon template created successfully", coupon)
}

// IssueCoupon 处理用户领取优惠券的 HTTP 请求。
func (h *Handler) IssueCoupon(c *gin.Context) {
	var req struct {
		UserID   uint64 `json:"user_id" binding:"required"`
		CouponID uint64 `json:"coupon_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	userCoupon, err := h.app.AcquireCoupon(c.Request.Context(), req.UserID, req.CouponID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to issue coupon", "user_id", req.UserID, "coupon_id", req.CouponID, "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "coupon issued successfully", userCoupon)
}

// ListCoupons 分页获取优惠券模板列表。
func (h *Handler) ListCoupons(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	list, total, err := h.app.ListCoupons(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list coupons", "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "coupons listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListUserCoupons 获取特定用户的优惠券持有列表。
func (h *Handler) ListUserCoupons(c *gin.Context) {
	var (
		userID uint64
		err    error
	)
	if val := c.Query("user_id"); val != "" {
		userID, err = strconv.ParseUint(val, 10, 64)
		if err != nil {
			response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user_id format", "")
			return
		}
	}
	if userID == 0 {
		response.ErrorWithStatus(c, http.StatusBadRequest, "user_id is required", "")
		return
	}
	status := c.Query("status")
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	list, total, err := h.app.ListUserCoupons(c.Request.Context(), userID, status, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list user coupons", "user_id", userID, "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "user coupons listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateActivity 创建优惠券营销活动。
func (h *Handler) CreateActivity(c *gin.Context) {
	var req struct {
		Name        string    `json:"name" binding:"required"`
		Description string    `json:"description"`
		StartTime   time.Time `json:"start_time" binding:"required"`
		EndTime     time.Time `json:"end_time" binding:"required"`
		CouponIDs   []uint64  `json:"coupon_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	activity := domain.NewCouponActivity(req.Name, req.Description, req.StartTime, req.EndTime, req.CouponIDs)
	err := h.app.CreateActivity(c.Request.Context(), activity)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to create activity", "name", req.Name, "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "activity created successfully", activity)
}

// UseCoupon 提交使用优惠券的请求。
func (h *Handler) UseCoupon(c *gin.Context) {
	userCouponID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user coupon id", "")
		return
	}
	var req struct {
		UserID  uint64 `json:"user_id" binding:"required"`
		OrderID string `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	err = h.app.UseCoupon(c.Request.Context(), userCouponID, req.UserID, req.OrderID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to use coupon", "id", userCouponID, "user_id", req.UserID, "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "coupon used successfully", nil)
}

// RegisterRoutes 注册路由.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/coupons")
	{
		group.POST("", h.CreateCoupon)
		group.GET("", h.ListCoupons)
		group.POST("/issue", h.IssueCoupon)
	}

	userGroup := r.Group("/user-coupons")
	{
		userGroup.GET("", h.ListUserCoupons)
		userGroup.POST("/:id/use", h.UseCoupon)
	}

	activityGroup := r.Group("/coupon-activities")
	{
		activityGroup.POST("", h.CreateActivity)
	}
}

package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。
	"time"     // 导入时间包，用于时间解析。

	"github.com/wyfcoding/ecommerce/internal/coupon/application"   // 导入优惠券模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/coupon/domain/entity" // 导入优惠券模块的领域实体。
	"github.com/wyfcoding/ecommerce/pkg/response"                  // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了Coupon模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.CouponService // 依赖Coupon应用服务，处理核心业务逻辑。
	logger  *slog.Logger               // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Coupon HTTP Handler 实例。
func NewHandler(service *application.CouponService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateCoupon 处理创建优惠券的HTTP请求。
// Method: POST
// Path: /coupons
func (h *Handler) CreateCoupon(c *gin.Context) {
	// 定义请求体结构，用于接收优惠券的创建信息。
	var req struct {
		Name           string            `json:"name" binding:"required"`            // 优惠券名称，必填。
		Description    string            `json:"description"`                        // 描述，选填。
		Type           entity.CouponType `json:"type" binding:"required"`            // 优惠券类型，必填。
		DiscountAmount int64             `json:"discount_amount" binding:"required"` // 优惠金额，必填。
		MinOrderAmount int64             `json:"min_order_amount"`                   // 最低订单金额，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建优惠券。
	coupon, err := h.service.CreateCoupon(c.Request.Context(), req.Name, req.Description, req.Type, req.DiscountAmount, req.MinOrderAmount)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create coupon", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create coupon", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Coupon created successfully", coupon)
}

// ActivateCoupon 处理激活优惠券的HTTP请求。
// Method: PUT
// Path: /coupons/:id/activate
func (h *Handler) ActivateCoupon(c *gin.Context) {
	// 从URL路径中解析优惠券ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层激活优惠券。
	err = h.service.ActivateCoupon(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to activate coupon", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to activate coupon", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Coupon activated successfully", nil)
}

// IssueCoupon 处理发放优惠券给用户的HTTP请求。
// Method: POST
// Path: /coupons/issue
func (h *Handler) IssueCoupon(c *gin.Context) {
	// 定义请求体结构，用于接收用户ID和优惠券ID。
	var req struct {
		UserID   uint64 `json:"user_id" binding:"required"`   // 用户ID，必填。
		CouponID uint64 `json:"coupon_id" binding:"required"` // 优惠券ID，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层发放优惠券。
	userCoupon, err := h.service.IssueCoupon(c.Request.Context(), req.UserID, req.CouponID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to issue coupon", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to issue coupon", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Coupon issued successfully", userCoupon)
}

// ListCoupons 处理获取优惠券列表的HTTP请求。
// Method: GET
// Path: /coupons
func (h *Handler) ListCoupons(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取优惠券列表。
	// TODO: 应用服务层的ListCoupons方法支持status过滤，此处未从请求参数获取。
	list, total, err := h.service.ListCoupons(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list coupons", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list coupons", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Coupons listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListUserCoupons 处理获取用户优惠券列表的HTTP请求。
// Method: GET
// Path: /user-coupons
func (h *Handler) ListUserCoupons(c *gin.Context) {
	// 从查询参数中获取用户ID。
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if userID == 0 {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid user ID", "user_id is required")
		return
	}
	// 从查询参数中获取状态过滤条件。
	status := c.Query("status")
	// 从查询参数中获取页码和每页大小。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取用户优惠券列表。
	list, total, err := h.service.ListUserCoupons(c.Request.Context(), userID, status, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list user coupons", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list user coupons", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "User coupons listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateActivity 处理创建优惠券活动的HTTP请求。
// Method: POST
// Path: /coupon-activities
func (h *Handler) CreateActivity(c *gin.Context) {
	// 定义请求体结构，用于接收优惠券活动的创建信息。
	var req struct {
		Name        string    `json:"name" binding:"required"`       // 活动名称，必填。
		Description string    `json:"description"`                   // 描述，选填。
		StartTime   time.Time `json:"start_time" binding:"required"` // 开始时间，必填。
		EndTime     time.Time `json:"end_time" binding:"required"`   // 结束时间，必填。
		CouponIDs   []uint64  `json:"coupon_ids" binding:"required"` // 关联优惠券ID列表，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建优惠券活动。
	activity, err := h.service.CreateActivity(c.Request.Context(), req.Name, req.Description, req.StartTime, req.EndTime, req.CouponIDs)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create activity", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create activity", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Activity created successfully", activity)
}

// RegisterRoutes 在给定的Gin路由组中注册Coupon模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /coupons 路由组，用于优惠券管理。
	group := r.Group("/coupons")
	{
		group.POST("", h.CreateCoupon)               // 创建优惠券。
		group.PUT("/:id/activate", h.ActivateCoupon) // 激活优惠券。
		group.GET("", h.ListCoupons)                 // 获取优惠券列表。
		group.POST("/issue", h.IssueCoupon)          // 发放优惠券。
		// TODO: 补充获取优惠券详情、更新、删除优惠券等接口。
	}

	// /user-coupons 路由组，用于用户优惠券管理。
	userGroup := r.Group("/user-coupons")
	{
		userGroup.GET("", h.ListUserCoupons) // 获取用户优惠券列表。
		// TODO: 补充使用优惠券、作废优惠券等接口。
	}

	// /coupon-activities 路由组，用于优惠券活动管理。
	activityGroup := r.Group("/coupon-activities")
	{
		activityGroup.POST("", h.CreateActivity) // 创建优惠券活动。
		// TODO: 补充获取、更新、删除活动等接口。
	}
}

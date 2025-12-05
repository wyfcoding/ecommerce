package http

import (
	"log/slog" // 导入结构化日志库。
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/gin-gonic/gin"                                         // 导入Gin Web框架。
	"github.com/wyfcoding/ecommerce/internal/subscription/application" // 导入订阅模块的应用服务。
	// "github.com/wyfcoding/ecommerce/pkg/response" // 未直接使用pkg/response，而是直接使用c.JSON。
)

// Handler 结构体定义了Subscription模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	app    *application.SubscriptionService // 依赖Subscription应用服务，处理核心业务逻辑。
	logger *slog.Logger                     // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Subscription HTTP Handler 实例。
func NewHandler(app *application.SubscriptionService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// RegisterRoutes 在给定的Gin路由组中注册Subscription模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /plans 路由组，用于订阅计划管理。
	plans := r.Group("/plans")
	{
		plans.POST("", h.CreatePlan)    // 创建订阅计划。
		plans.GET("", h.ListPlans)      // 获取订阅计划列表。
		plans.GET("/:id", h.GetPlan)    // 获取计划详情。
		plans.PUT("/:id", h.UpdatePlan) // 更新计划。
	}

	// /subscriptions 路由组，用于用户订阅管理。
	subscriptions := r.Group("/subscriptions")
	{
		subscriptions.POST("", h.Subscribe)          // 订阅计划。
		subscriptions.GET("", h.ListSubscriptions)   // 获取用户订阅列表。
		subscriptions.GET("/:id", h.GetSubscription) // 获取订阅详情。
		subscriptions.POST("/:id/cancel", h.Cancel)  // 取消订阅。
		subscriptions.POST("/:id/renew", h.Renew)    // 续订。
	}
}

// CreatePlan 处理创建订阅计划的HTTP请求。
// Method: POST
// Path: /plans
func (h *Handler) CreatePlan(c *gin.Context) {
	// 定义请求体结构，用于接收订阅计划的创建信息。
	var req struct {
		Name        string   `json:"name"`        // 计划名称。
		Description string   `json:"description"` // 计划描述。
		Price       uint64   `json:"price"`       // 价格。
		Duration    int32    `json:"duration"`    // 时长。
		Features    []string `json:"features"`    // 特性列表。
	}
	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用应用服务层创建计划。
	plan, err := h.app.CreatePlan(c.Request.Context(), req.Name, req.Description, req.Price, req.Duration, req.Features)
	if err != nil {
		h.logger.Error("failed to create plan", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回成功的响应。
	c.JSON(http.StatusOK, plan)
}

// ListPlans 处理获取订阅计划列表的HTTP请求。
// Method: GET
// Path: /plans
func (h *Handler) ListPlans(c *gin.Context) {
	// 调用应用服务层获取订阅计划列表。
	plans, err := h.app.ListPlans(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list plans", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回成功的响应。
	c.JSON(http.StatusOK, plans)
}

// Subscribe 处理用户订阅的HTTP请求。
// Method: POST
// Path: /subscriptions
func (h *Handler) Subscribe(c *gin.Context) {
	// 定义请求体结构，用于接收用户ID和计划ID。
	var req struct {
		UserID uint64 `json:"user_id"` // 用户ID。
		PlanID uint64 `json:"plan_id"` // 计划ID。
	}
	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用应用服务层执行订阅。
	sub, err := h.app.Subscribe(c.Request.Context(), req.UserID, req.PlanID)
	if err != nil {
		h.logger.Error("failed to subscribe", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回成功的响应。
	c.JSON(http.StatusOK, sub)
}

// Cancel 处理取消订阅的HTTP请求。
// Method: POST
// Path: /subscriptions/:id/cancel
func (h *Handler) Cancel(c *gin.Context) {
	// 从URL路径中解析订阅ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 调用应用服务层取消订阅。
	if err := h.app.Cancel(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to cancel subscription", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回成功的响应。
	c.JSON(http.StatusOK, gin.H{"status": "canceled"})
}

// Renew 处理续订的HTTP请求。
// Method: POST
// Path: /subscriptions/:id/renew
func (h *Handler) Renew(c *gin.Context) {
	// 从URL路径中解析订阅ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 调用应用服务层续订。
	if err := h.app.Renew(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to renew subscription", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回成功的响应。
	c.JSON(http.StatusOK, gin.H{"status": "renewed"})
}

// ListSubscriptions 处理获取用户订阅列表的HTTP请求。
// Method: GET
// Path: /subscriptions
func (h *Handler) ListSubscriptions(c *gin.Context) {
	// 从查询参数中获取用户ID。
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取用户订阅列表。
	subs, total, err := h.app.ListSubscriptions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.Error("failed to list subscriptions", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回包含分页信息的成功响应。
	c.JSON(http.StatusOK, gin.H{
		"items":       subs,
		"total_count": total,
	})
}

// GetPlan handles the request to get a subscription plan by ID.
// Method: GET
// Path: /plans/:id
func (h *Handler) GetPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	plan, err := h.app.GetPlan(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get plan", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if plan == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}

	c.JSON(http.StatusOK, plan)
}

// UpdatePlan handles the request to update a subscription plan.
// Method: PUT
// Path: /plans/:id
func (h *Handler) UpdatePlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Price       *uint64  `json:"price"`
		Duration    *int32   `json:"duration"`
		Features    []string `json:"features"`
		Enabled     *bool    `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plan, err := h.app.UpdatePlan(c.Request.Context(), id, req.Name, req.Description, req.Price, req.Duration, req.Features, req.Enabled)
	if err != nil {
		h.logger.Error("failed to update plan", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plan)
}

// GetSubscription handles the request to get a subscription by ID.
// Method: GET
// Path: /subscriptions/:id
func (h *Handler) GetSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	sub, err := h.app.GetSubscription(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get subscription", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sub == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

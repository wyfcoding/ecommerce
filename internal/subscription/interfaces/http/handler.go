package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/subscription/application"
)

// Handler 结构体定义了Subscription模块的HTTP处理层。
type Handler struct {
	app    *application.SubscriptionService
	logger *slog.Logger
}

// NewHandler 创建并返回一个新的 Subscription HTTP Handler 实例。
func NewHandler(app *application.SubscriptionService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// RegisterRoutes 在给定的Gin路由组中注册Subscription模块的HTTP路由。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /plans 路由组，用于订阅计划管理。
	plans := r.Group("/plans")
	{
		plans.POST("", h.CreatePlan)
		plans.GET("", h.ListPlans)
		plans.GET("/:id", h.GetPlan)
		plans.PUT("/:id", h.UpdatePlan)
	}

	// /subscriptions 路由组，用于用户订阅管理。
	subscriptions := r.Group("/subscriptions")
	{
		subscriptions.POST("", h.Subscribe)
		subscriptions.GET("", h.ListSubscriptions)
		subscriptions.GET("/:id", h.GetSubscription)
		subscriptions.POST("/:id/cancel", h.Cancel)
		subscriptions.POST("/:id/renew", h.Renew)
	}
}

// CreatePlan 处理创建订阅计划的HTTP请求。
func (h *Handler) CreatePlan(c *gin.Context) {
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Price       uint64   `json:"price"`
		Duration    int32    `json:"duration"`
		Features    []string `json:"features"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	plan, err := h.app.CreatePlan(c.Request.Context(), req.Name, req.Description, req.Price, req.Duration, req.Features)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to create plan", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, plan)
}

// ListPlans 处理获取订阅计划列表的HTTP请求。
func (h *Handler) ListPlans(c *gin.Context) {
	plans, err := h.app.ListPlans(c.Request.Context())
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list plans", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, plans)
}

// Subscribe 处理用户订阅的HTTP请求。
func (h *Handler) Subscribe(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id"`
		PlanID uint64 `json:"plan_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	sub, err := h.app.Subscribe(c.Request.Context(), req.UserID, req.PlanID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to subscribe", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, sub)
}

// Cancel 处理取消订阅的HTTP请求。
func (h *Handler) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id", "")
		return
	}

	if err := h.app.Cancel(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to cancel subscription", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, gin.H{"status": "canceled"})
}

// Renew 处理续订的HTTP请求。
func (h *Handler) Renew(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id", "")
		return
	}

	if err := h.app.Renew(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to renew subscription", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, gin.H{"status": "renewed"})
}

// ListSubscriptions 处理获取用户订阅列表的HTTP请求。
func (h *Handler) ListSubscriptions(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user_id", "")
		return
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	subs, total, err := h.app.ListSubscriptions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list subscriptions", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":       subs,
		"total_count": total,
	})
}

// GetPlan 处理通过 ID 获取订阅计划的请求。
func (h *Handler) GetPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id", "")
		return
	}

	plan, err := h.app.GetPlan(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get plan", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}
	if plan == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "plan not found", "")
		return
	}

	response.Success(c, plan)
}

// UpdatePlan 处理更新订阅计划的请求。
func (h *Handler) UpdatePlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id", "")
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
		response.ErrorWithStatus(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	plan, err := h.app.UpdatePlan(c.Request.Context(), id, req.Name, req.Description, req.Price, req.Duration, req.Features, req.Enabled)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to update plan", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, plan)
}

// GetSubscription 处理通过 ID 获取订阅信息的请求。
func (h *Handler) GetSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id", "")
		return
	}

	sub, err := h.app.GetSubscription(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get subscription", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}
	if sub == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "subscription not found", "")
		return
	}

	response.Success(c, sub)
}

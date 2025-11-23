package http

import (
	"net/http"
	"strconv"

	"ecommerce/internal/subscription/application"
	"ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.SubscriptionService
	logger  *slog.Logger
}

func NewHandler(service *application.SubscriptionService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreatePlan 创建计划
func (h *Handler) CreatePlan(c *gin.Context) {
	var req struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Price       uint64   `json:"price" binding:"required"`
		Duration    int32    `json:"duration" binding:"required"`
		Features    []string `json:"features"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	plan, err := h.service.CreatePlan(c.Request.Context(), req.Name, req.Description, req.Price, req.Duration, req.Features)
	if err != nil {
		h.logger.Error("Failed to create plan", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create plan", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Plan created successfully", plan)
}

// Subscribe 订阅
func (h *Handler) Subscribe(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
		PlanID uint64 `json:"plan_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	sub, err := h.service.Subscribe(c.Request.Context(), req.UserID, req.PlanID)
	if err != nil {
		h.logger.Error("Failed to subscribe", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to subscribe", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Subscribed successfully", sub)
}

// Cancel 取消
func (h *Handler) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.Cancel(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to cancel subscription", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to cancel subscription", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Subscription canceled successfully", nil)
}

// Renew 续订
func (h *Handler) Renew(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.Renew(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to renew subscription", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to renew subscription", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Subscription renewed successfully", nil)
}

// ListPlans 计划列表
func (h *Handler) ListPlans(c *gin.Context) {
	plans, err := h.service.ListPlans(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list plans", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list plans", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Plans listed successfully", plans)
}

// ListSubscriptions 订阅列表
func (h *Handler) ListSubscriptions(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListSubscriptions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list subscriptions", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list subscriptions", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Subscriptions listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/subscription")
	{
		group.POST("/plans", h.CreatePlan)
		group.GET("/plans", h.ListPlans)
		group.POST("/subscribe", h.Subscribe)
		group.POST("/:id/cancel", h.Cancel)
		group.POST("/:id/renew", h.Renew)
		group.GET("", h.ListSubscriptions)
	}
}

package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/subscription/application"
)

type Handler struct {
	app    *application.SubscriptionService
	logger *slog.Logger
}

func NewHandler(app *application.SubscriptionService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	plans := r.Group("/plans")
	{
		plans.POST("", h.CreatePlan)
		plans.GET("", h.ListPlans)
	}

	subscriptions := r.Group("/subscriptions")
	{
		subscriptions.POST("", h.Subscribe)
		subscriptions.GET("", h.ListSubscriptions)
		subscriptions.POST("/:id/cancel", h.Cancel)
		subscriptions.POST("/:id/renew", h.Renew)
	}
}

func (h *Handler) CreatePlan(c *gin.Context) {
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Price       uint64   `json:"price"`
		Duration    int32    `json:"duration"`
		Features    []string `json:"features"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plan, err := h.app.CreatePlan(c.Request.Context(), req.Name, req.Description, req.Price, req.Duration, req.Features)
	if err != nil {
		h.logger.Error("failed to create plan", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plan)
}

func (h *Handler) ListPlans(c *gin.Context) {
	plans, err := h.app.ListPlans(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list plans", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plans)
}

func (h *Handler) Subscribe(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id"`
		PlanID uint64 `json:"plan_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub, err := h.app.Subscribe(c.Request.Context(), req.UserID, req.PlanID)
	if err != nil {
		h.logger.Error("failed to subscribe", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sub)
}

func (h *Handler) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.app.Cancel(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to cancel subscription", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "canceled"})
}

func (h *Handler) Renew(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.app.Renew(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to renew subscription", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "renewed"})
}

func (h *Handler) ListSubscriptions(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	subs, total, err := h.app.ListSubscriptions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.Error("failed to list subscriptions", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":       subs,
		"total_count": total,
	})
}

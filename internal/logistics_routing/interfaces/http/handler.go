package http

import (
	"net/http"

	"github.com/wyfcoding/ecommerce/internal/logistics_routing/application"
	"github.com/wyfcoding/ecommerce/internal/logistics_routing/domain"
	"github.com/wyfcoding/pkg/response"

	"log/slog"

	"github.com/gin-gonic/gin"
)

// Handler 结构体定义了LogisticsRouting模块的HTTP处理层。
type Handler struct {
	app    *application.LogisticsRoutingService
	logger *slog.Logger
}

// NewHandler 创建并返回一个新的 LogisticsRouting HTTP Handler 实例。
func NewHandler(app *application.LogisticsRoutingService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// RegisterCarrier 处理注册配送商的HTTP请求。
func (h *Handler) RegisterCarrier(c *gin.Context) {
	var req domain.Carrier
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 确保默认值
	req.Rating = 5.0
	req.IsActive = true

	if err := h.app.RegisterCarrier(c.Request.Context(), &req); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to register carrier", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to register carrier", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Carrier registered successfully", req)
}

// OptimizeRoute 处理优化配送路线的HTTP请求。
func (h *Handler) OptimizeRoute(c *gin.Context) {
	var req struct {
		OrderIDs []uint64 `json:"order_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	route, err := h.app.OptimizeRoute(c.Request.Context(), req.OrderIDs)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to optimize route", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to optimize route", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Route optimized successfully", route)
}

// ListCarriers 处理获取配送商列表的HTTP请求。
func (h *Handler) ListCarriers(c *gin.Context) {
	carriers, err := h.app.ListCarriers(c.Request.Context())
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list carriers", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list carriers", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Carriers listed successfully", carriers)
}

// RegisterRoutes 在给定的Gin路由组中注册LogisticsRouting模块的HTTP路由。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/logistics-routing")
	{
		group.POST("/carriers", h.RegisterCarrier)
		group.GET("/carriers", h.ListCarriers)
		group.POST("/optimize", h.OptimizeRoute)
	}
}

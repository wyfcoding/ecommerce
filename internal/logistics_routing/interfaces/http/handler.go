package http

import (
	"net/http"

	"github.com/wyfcoding/ecommerce/internal/logistics_routing/application"
	"github.com/wyfcoding/ecommerce/internal/logistics_routing/domain/entity"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.LogisticsRoutingService
	logger  *slog.Logger
}

func NewHandler(service *application.LogisticsRoutingService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterCarrier 注册配送商
func (h *Handler) RegisterCarrier(c *gin.Context) {
	var req entity.Carrier
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.RegisterCarrier(c.Request.Context(), &req); err != nil {
		h.logger.Error("Failed to register carrier", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to register carrier", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Carrier registered successfully", req)
}

// OptimizeRoute 优化路由
func (h *Handler) OptimizeRoute(c *gin.Context) {
	var req struct {
		OrderIDs []uint64 `json:"order_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	route, err := h.service.OptimizeRoute(c.Request.Context(), req.OrderIDs)
	if err != nil {
		h.logger.Error("Failed to optimize route", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to optimize route", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Route optimized successfully", route)
}

// ListCarriers 获取配送商列表
func (h *Handler) ListCarriers(c *gin.Context) {
	carriers, err := h.service.ListCarriers(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list carriers", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list carriers", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Carriers listed successfully", carriers)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/logistics-routing")
	{
		group.POST("/carriers", h.RegisterCarrier)
		group.GET("/carriers", h.ListCarriers)
		group.POST("/optimize", h.OptimizeRoute)
	}
}

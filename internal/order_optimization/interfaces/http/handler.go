package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/order_optimization/application"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.OrderOptimizationService
	logger  *slog.Logger
}

func NewHandler(service *application.OrderOptimizationService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// MergeOrders 合并订单
func (h *Handler) MergeOrders(c *gin.Context) {
	var req struct {
		UserID   uint64   `json:"user_id" binding:"required"`
		OrderIDs []uint64 `json:"order_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	mergedOrder, err := h.service.MergeOrders(c.Request.Context(), req.UserID, req.OrderIDs)
	if err != nil {
		h.logger.Error("Failed to merge orders", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to merge orders", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Orders merged successfully", mergedOrder)
}

// SplitOrder 拆分订单
func (h *Handler) SplitOrder(c *gin.Context) {
	var req struct {
		OrderID uint64 `json:"order_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	splitOrders, err := h.service.SplitOrder(c.Request.Context(), req.OrderID)
	if err != nil {
		h.logger.Error("Failed to split order", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to split order", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Order split successfully", splitOrders)
}

// AllocateWarehouse 分配仓库
func (h *Handler) AllocateWarehouse(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Order ID", err.Error())
		return
	}

	plan, err := h.service.AllocateWarehouse(c.Request.Context(), orderID)
	if err != nil {
		h.logger.Error("Failed to allocate warehouse", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to allocate warehouse", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Warehouse allocated successfully", plan)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/order-optimization")
	{
		group.POST("/merge", h.MergeOrders)
		group.POST("/split", h.SplitOrder)
		group.POST("/allocations/:order_id", h.AllocateWarehouse)
	}
}

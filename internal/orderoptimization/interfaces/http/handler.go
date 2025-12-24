package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/orderoptimization/application"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 处理 HTTP 或 gRPC 请求。
type Handler struct {
	service *application.OrderOptimizationService
	logger  *slog.Logger
}

// NewHandler 处理 HTTP 或 gRPC 请求。
func NewHandler(service *application.OrderOptimizationService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

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

func (h *Handler) GetMergedOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	order, err := h.service.GetMergedOrder(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get merged order", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get merged order", err.Error())
		return
	}
	if order == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "Merged order not found", "order not found")
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Merged order retrieved successfully", order)
}

func (h *Handler) ListSplitOrders(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Order ID", err.Error())
		return
	}

	orders, err := h.service.ListSplitOrders(c.Request.Context(), orderID)
	if err != nil {
		h.logger.Error("Failed to list split orders", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list split orders", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Split orders retrieved successfully", orders)
}

func (h *Handler) GetAllocationPlan(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Order ID", err.Error())
		return
	}

	plan, err := h.service.GetAllocationPlan(c.Request.Context(), orderID)
	if err != nil {
		h.logger.Error("Failed to get allocation plan", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get allocation plan", err.Error())
		return
	}
	if plan == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "Allocation plan not found", "plan not found")
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Allocation plan retrieved successfully", plan)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/order-optimization")
	{
		group.POST("/merge", h.MergeOrders)
		group.GET("/merge/:id", h.GetMergedOrder)
		group.POST("/split", h.SplitOrder)
		group.GET("/split/:order_id", h.ListSplitOrders)
		group.POST("/allocations/:order_id", h.AllocateWarehouse)
		group.GET("/allocations/:order_id", h.GetAllocationPlan)
	}
}

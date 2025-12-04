package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/order_optimization/application" // 导入订单优化模块的应用服务。
	"github.com/wyfcoding/ecommerce/pkg/response"                            // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了OrderOptimization模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.OrderOptimizationService // 依赖OrderOptimization应用服务，处理核心业务逻辑。
	logger  *slog.Logger                          // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 OrderOptimization HTTP Handler 实例。
func NewHandler(service *application.OrderOptimizationService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// MergeOrders 处理合并订单的HTTP请求。
// Method: POST
// Path: /order-optimization/merge
func (h *Handler) MergeOrders(c *gin.Context) {
	// 定义请求体结构，用于接收待合并订单的用户ID和订单ID列表。
	var req struct {
		UserID   uint64   `json:"user_id" binding:"required"`   // 用户ID，必填。
		OrderIDs []uint64 `json:"order_ids" binding:"required"` // 待合并的原始订单ID列表，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层合并订单。
	mergedOrder, err := h.service.MergeOrders(c.Request.Context(), req.UserID, req.OrderIDs)
	if err != nil {
		h.logger.Error("Failed to merge orders", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to merge orders", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created，包含合并后的订单信息。
	response.SuccessWithStatus(c, http.StatusCreated, "Orders merged successfully", mergedOrder)
}

// SplitOrder 处理拆分订单的HTTP请求。
// Method: POST
// Path: /order-optimization/split
func (h *Handler) SplitOrder(c *gin.Context) {
	// 定义请求体结构，用于接收待拆分的原始订单ID。
	var req struct {
		OrderID uint64 `json:"order_id" binding:"required"` // 原始订单ID，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层拆分订单。
	splitOrders, err := h.service.SplitOrder(c.Request.Context(), req.OrderID)
	if err != nil {
		h.logger.Error("Failed to split order", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to split order", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created，包含拆分后的子订单列表。
	response.SuccessWithStatus(c, http.StatusCreated, "Order split successfully", splitOrders)
}

// AllocateWarehouse 处理仓库分配的HTTP请求。
// Method: POST
// Path: /order-optimization/allocations/:order_id
func (h *Handler) AllocateWarehouse(c *gin.Context) {
	// 从URL路径中解析订单ID。
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Order ID", err.Error())
		return
	}

	// 调用应用服务层分配仓库。
	plan, err := h.service.AllocateWarehouse(c.Request.Context(), orderID)
	if err != nil {
		h.logger.Error("Failed to allocate warehouse", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to allocate warehouse", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created，包含仓库分配计划。
	response.SuccessWithStatus(c, http.StatusCreated, "Warehouse allocated successfully", plan)
}

// RegisterRoutes 在给定的Gin路由组中注册OrderOptimization模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /order-optimization 路由组，用于所有订单优化相关接口。
	group := r.Group("/order-optimization")
	{
		group.POST("/merge", h.MergeOrders)                       // 合并订单。
		group.POST("/split", h.SplitOrder)                        // 拆分订单。
		group.POST("/allocations/:order_id", h.AllocateWarehouse) // 分配仓库。
		// TODO: 补充获取合并订单详情、获取拆分订单列表、获取仓库分配计划详情等接口。
	}
}

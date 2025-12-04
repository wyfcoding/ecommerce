package http

import (
	"net/http" // 导入HTTP状态码。

	"github.com/wyfcoding/ecommerce/internal/logistics_routing/application"   // 导入物流路由模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/logistics_routing/domain/entity" // 导入物流路由模块的领域实体。
	"github.com/wyfcoding/ecommerce/pkg/response"                             // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了LogisticsRouting模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.LogisticsRoutingService // 依赖LogisticsRouting应用服务，处理核心业务逻辑。
	logger  *slog.Logger                         // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 LogisticsRouting HTTP Handler 实例。
func NewHandler(service *application.LogisticsRoutingService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterCarrier 处理注册配送商的HTTP请求。
// Method: POST
// Path: /logistics-routing/carriers
func (h *Handler) RegisterCarrier(c *gin.Context) {
	// 定义请求体结构，使用 entity.Carrier 结构体直接绑定。
	var req entity.Carrier
	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层注册配送商。
	if err := h.service.RegisterCarrier(c.Request.Context(), &req); err != nil {
		h.logger.Error("Failed to register carrier", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to register carrier", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Carrier registered successfully", req)
}

// OptimizeRoute 处理优化配送路线的HTTP请求。
// Method: POST
// Path: /logistics-routing/optimize
func (h *Handler) OptimizeRoute(c *gin.Context) {
	// 定义请求体结构，用于接收订单ID列表。
	var req struct {
		OrderIDs []uint64 `json:"order_ids" binding:"required"` // 订单ID列表，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层优化配送路线。
	route, err := h.service.OptimizeRoute(c.Request.Context(), req.OrderIDs)
	if err != nil {
		h.logger.Error("Failed to optimize route", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to optimize route", err.Error())
		return
	}

	// 返回成功的响应，包含优化后的路线信息。
	response.SuccessWithStatus(c, http.StatusOK, "Route optimized successfully", route)
}

// ListCarriers 处理获取配送商列表的HTTP请求。
// Method: GET
// Path: /logistics-routing/carriers
func (h *Handler) ListCarriers(c *gin.Context) {
	// 调用应用服务层获取配送商列表。
	carriers, err := h.service.ListCarriers(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list carriers", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list carriers", err.Error())
		return
	}

	// 返回成功的响应，包含配送商列表。
	response.SuccessWithStatus(c, http.StatusOK, "Carriers listed successfully", carriers)
}

// RegisterRoutes 在给定的Gin路由组中注册LogisticsRouting模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /logistics-routing 路由组，用于所有物流路由相关接口。
	group := r.Group("/logistics-routing")
	{
		// 配送商管理接口。
		group.POST("/carriers", h.RegisterCarrier) // 注册配送商。
		group.GET("/carriers", h.ListCarriers)     // 获取配送商列表。
		// TODO: 补充获取配送商详情、更新配送商、删除配送商、激活/停用配送商的接口。

		// 路由优化接口。
		group.POST("/optimize", h.OptimizeRoute) // 优化配送路线。
		// TODO: 补充获取优化路线详情、列出优化路线等接口。
	}
}

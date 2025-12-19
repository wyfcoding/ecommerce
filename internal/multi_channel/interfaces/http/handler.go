package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/multi_channel/application"   // 导入多渠道模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/multi_channel/domain/entity" // 导入多渠道模块的领域实体。
	"github.com/wyfcoding/pkg/response"                                   // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了MultiChannel模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.MultiChannelService // 依赖MultiChannel应用服务，处理核心业务逻辑。
	logger  *slog.Logger                     // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 MultiChannel HTTP Handler 实例。
func NewHandler(service *application.MultiChannelService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterChannel 处理注册渠道的HTTP请求。
// Method: POST
// Path: /multi-channel/channels
func (h *Handler) RegisterChannel(c *gin.Context) {
	// 定义请求体结构，使用 entity.Channel 结构体直接绑定。
	var req entity.Channel
	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层注册渠道。
	if err := h.service.RegisterChannel(c.Request.Context(), &req); err != nil {
		h.logger.Error("Failed to register channel", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to register channel", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Channel registered successfully", req)
}

// SyncOrders 处理同步订单的HTTP请求。
// Method: POST
// Path: /multi-channel/sync/orders
func (h *Handler) SyncOrders(c *gin.Context) {
	// 定义请求体结构，用于接收渠道ID。
	var req struct {
		ChannelID uint64 `json:"channel_id" binding:"required"` // 渠道ID，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层同步订单。
	if err := h.service.SyncOrders(c.Request.Context(), req.ChannelID); err != nil {
		h.logger.Error("Failed to sync orders", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to sync orders", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Orders synced successfully", nil)
}

// ListChannels 处理获取渠道列表的HTTP请求。
// Method: GET
// Path: /multi-channel/channels
func (h *Handler) ListChannels(c *gin.Context) {
	// 调用应用服务层获取渠道列表。
	channels, err := h.service.ListChannels(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list channels", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list channels", err.Error())
		return
	}

	// 返回成功的响应，包含渠道列表。
	response.SuccessWithStatus(c, http.StatusOK, "Channels listed successfully", channels)
}

// ListOrders 处理获取订单列表的HTTP请求。
// Method: GET
// Path: /multi-channel/orders
func (h *Handler) ListOrders(c *gin.Context) {
	// 从查询参数中获取渠道ID、状态、页码和每页大小，并设置默认值。
	channelID, _ := strconv.ParseUint(c.Query("channel_id"), 10, 64)
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取订单列表。
	list, total, err := h.service.ListOrders(c.Request.Context(), channelID, status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list orders", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list orders", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Orders listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册MultiChannel模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /multi-channel 路由组，用于所有多渠道相关接口。
	group := r.Group("/multi-channel")
	{
		// 渠道管理接口。
		group.POST("/channels", h.RegisterChannel) // 注册渠道。
		group.GET("/channels", h.ListChannels)     // 获取渠道列表。
		// TODO: 补充获取渠道详情、更新渠道、删除渠道的接口。

		// 订单同步接口。
		group.POST("/sync/orders", h.SyncOrders) // 同步订单。
		group.GET("/orders", h.ListOrders)       // 获取同步后的订单列表。
		// TODO: 补充获取订单详情等接口。
	}
}

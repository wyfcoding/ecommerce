package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/user_tier/application" // 导入用户等级模块的应用服务。
	"github.com/wyfcoding/ecommerce/pkg/response"                   // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了UserTier模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.UserTierService // 依赖UserTier应用服务，处理核心业务逻辑。
	logger  *slog.Logger                 // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 UserTier HTTP Handler 实例。
func NewHandler(service *application.UserTierService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetTier 处理获取用户等级信息的HTTP请求。
// Method: GET
// Path: /user_tier/:user_id
func (h *Handler) GetTier(c *gin.Context) {
	// 从URL路径中解析用户ID。
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	// 调用应用服务层获取用户等级。
	tier, err := h.service.GetUserTier(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user tier", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get user tier", err.Error())
		return
	}

	// 返回成功的响应，包含用户等级信息。
	response.SuccessWithStatus(c, http.StatusOK, "User tier retrieved successfully", tier)
}

// GetPoints 处理获取用户积分余额的HTTP请求。
// Method: GET
// Path: /user_tier/:user_id/points
func (h *Handler) GetPoints(c *gin.Context) {
	// 从URL路径中解析用户ID。
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	// 调用应用服务层获取用户积分。
	points, err := h.service.GetPoints(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get points", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get points", err.Error())
		return
	}

	// 返回成功的响应，包含积分余额。
	response.SuccessWithStatus(c, http.StatusOK, "Points retrieved successfully", gin.H{"points": points})
}

// Exchange 处理兑换商品的HTTP请求。
// Method: POST
// Path: /user_tier/exchange
func (h *Handler) Exchange(c *gin.Context) {
	// 定义请求体结构，用于接收兑换商品的详细信息。
	var req struct {
		UserID     uint64 `json:"user_id" binding:"required"`     // 用户ID，必填。
		ExchangeID uint64 `json:"exchange_id" binding:"required"` // 兑换商品ID，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层执行兑换。
	if err := h.service.Exchange(c.Request.Context(), req.UserID, req.ExchangeID); err != nil {
		h.logger.Error("Failed to exchange", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to exchange", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Exchange successful", nil)
}

// ListExchanges 处理获取可兑换商品列表的HTTP请求。
// Method: GET
// Path: /user_tier/exchanges
func (h *Handler) ListExchanges(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取可兑换商品列表。
	list, total, err := h.service.ListExchanges(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list exchanges", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list exchanges", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Exchanges listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListPointsLogs 处理获取用户积分日志列表的HTTP请求。
// Method: GET
// Path: /user_tier/:user_id/points/logs
func (h *Handler) ListPointsLogs(c *gin.Context) {
	// 从URL路径中解析用户ID。
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取积分日志列表。
	list, total, err := h.service.ListPointsLogs(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list points logs", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list points logs", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Points logs listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册UserTier模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /user_tier 路由组，用于所有用户等级和积分相关接口。
	group := r.Group("/user_tier")
	{
		group.GET("/:user_id", h.GetTier)                    // 获取用户等级信息。
		group.GET("/:user_id/points", h.GetPoints)           // 获取用户积分余额。
		group.GET("/:user_id/points/logs", h.ListPointsLogs) // 获取用户积分日志列表。
		group.POST("/exchange", h.Exchange)                  // 兑换商品。
		group.GET("/exchanges", h.ListExchanges)             // 获取可兑换商品列表。
		// TODO: 补充增加/扣除成长值、增加/扣除积分、获取等级配置等接口。
	}
}

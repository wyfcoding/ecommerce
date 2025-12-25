package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/gateway/application" // 导入网关模块的应用服务。
	"github.com/wyfcoding/pkg/response"                           // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了Gateway模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.Gateway // 依赖Gateway应用服务，处理核心业务逻辑。
	logger  *slog.Logger                // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Gateway HTTP Handler 实例。
func NewHandler(service *application.Gateway, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoute 处理注册API路由的HTTP请求。
// HTTP 方法: POST
// 请求路径: /gateway/routes
func (h *Handler) RegisterRoute(c *gin.Context) {
	// 定义请求体结构，用于接收路由注册信息。
	var req struct {
		Path        string `json:"path" binding:"required"`    // 路由路径，必填。
		Method      string `json:"method" binding:"required"`  // HTTP方法，必填。
		Service     string `json:"service" binding:"required"` // 目标服务名，必填。
		Backend     string `json:"backend" binding:"required"` // 后端地址，必填。
		Timeout     int32  `json:"timeout"`                    // 超时时间（毫秒），选填。
		Retries     int32  `json:"retries"`                    // 重试次数，选填。
		Description string `json:"description"`                // 描述，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层注册路由。
	route, err := h.service.RegisterRoute(c.Request.Context(), req.Path, req.Method, req.Service, req.Backend, req.Timeout, req.Retries, req.Description)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to register route", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to register route", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Route registered successfully", route)
}

// ListRoutes 处理获取路由列表的HTTP请求。
// HTTP 方法: GET
// 请求路径: /gateway/routes
func (h *Handler) ListRoutes(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取路由列表。
	list, total, err := h.service.ListRoutes(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list routes", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list routes", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Routes listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// DeleteRoute 处理删除API路由的HTTP请求。
// HTTP 方法: DELETE
// 请求路径: /gateway/routes/:id
func (h *Handler) DeleteRoute(c *gin.Context) {
	// 从URL路径中解析路由ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层删除路由。
	if err := h.service.DeleteRoute(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to delete route", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete route", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Route deleted successfully", nil)
}

// AddRateLimitRule 处理添加限流规则的HTTP请求。
// HTTP 方法: POST
// 请求路径: /gateway/ratelimits
func (h *Handler) AddRateLimitRule(c *gin.Context) {
	// 定义请求体结构，用于接收限流规则信息。
	var req struct {
		Name        string `json:"name" binding:"required"`   // 规则名称，必填。
		Path        string `json:"path" binding:"required"`   // 匹配路径，必填。
		Method      string `json:"method" binding:"required"` // HTTP方法，必填。
		Limit       int32  `json:"limit" binding:"required"`  // 限制请求数，必填。
		Window      int32  `json:"window" binding:"required"` // 时间窗口（秒），必填。
		Description string `json:"description"`               // 描述，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层添加限流规则。
	rule, err := h.service.AddRateLimitRule(c.Request.Context(), req.Name, req.Path, req.Method, req.Limit, req.Window, req.Description)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to add rate limit rule", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add rate limit rule", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Rate limit rule added successfully", rule)
}

// ListRateLimitRules 处理获取限流规则列表的HTTP请求。
// HTTP 方法: GET
// 请求路径: /gateway/ratelimits
func (h *Handler) ListRateLimitRules(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取限流规则列表。
	list, total, err := h.service.ListRateLimitRules(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list rate limit rules", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list rate limit rules", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Rate limit rules listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// DeleteRateLimitRule 处理删除限流规则的HTTP请求。
// HTTP 方法: DELETE
// 请求路径: /gateway/ratelimits/:id
func (h *Handler) DeleteRateLimitRule(c *gin.Context) {
	// 从URL路径中解析限流规则ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层删除限流规则。
	if err := h.service.DeleteRateLimitRule(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to delete rate limit rule", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete rate limit rule", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Rate limit rule deleted successfully", nil)
}

// RegisterRoutes 在给定的Gin路由组中注册Gateway模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /gateway 路由组，用于所有网关配置相关接口。
	group := r.Group("/gateway")
	{
		// 路由管理接口。
		group.POST("/routes", h.RegisterRoute)     // 注册路由。
		group.GET("/routes", h.ListRoutes)         // 获取路由列表。
		group.DELETE("/routes/:id", h.DeleteRoute) // 删除路由。
		// TODO: 补充获取路由详情、更新路由的接口。

		// 限流规则管理接口。
		group.POST("/ratelimits", h.AddRateLimitRule)          // 添加限流规则。
		group.GET("/ratelimits", h.ListRateLimitRules)         // 获取限流规则列表。
		group.DELETE("/ratelimits/:id", h.DeleteRateLimitRule) // 删除限流规则。
		// TODO: 补充获取限流规则详情、更新限流规则的接口。
	}
}

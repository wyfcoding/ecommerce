package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。
	"time"     // 导入时间包，用于时间解析。

	"github.com/wyfcoding/ecommerce/internal/audit/application"       // 导入审计模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/audit/domain/entity"     // 导入审计模块的领域实体。
	"github.com/wyfcoding/ecommerce/internal/audit/domain/repository" // 导入审计模块的领域仓储查询对象。
	"github.com/wyfcoding/ecommerce/pkg/response"                     // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了Audit模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.AuditService // 依赖Audit应用服务，处理核心业务逻辑。
	logger  *slog.Logger              // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Audit HTTP Handler 实例。
func NewHandler(service *application.AuditService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// QueryLogs 处理查询审计日志的HTTP请求。
// Method: GET
// Path: /audit/logs
func (h *Handler) QueryLogs(c *gin.Context) {
	// 从查询参数中获取分页和过滤条件，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userID, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)
	eventType := c.DefaultQuery("event_type", "")
	module := c.DefaultQuery("module", "")
	resourceType := c.DefaultQuery("resource_type", "")
	startTimeStr := c.DefaultQuery("start_time", "")
	endTimeStr := c.DefaultQuery("end_time", "")

	var startTime, endTime time.Time
	// 解析起始时间字符串。
	if startTimeStr != "" {
		startTime, _ = time.Parse(time.RFC3339, startTimeStr)
	}
	// 解析结束时间字符串。
	if endTimeStr != "" {
		endTime, _ = time.Parse(time.RFC3339, endTimeStr)
	}

	// 构建查询对象。
	query := &repository.AuditLogQuery{
		UserID:       userID,
		EventType:    entity.AuditEventType(eventType), // 将字符串转换为实体枚举类型。
		Module:       module,
		ResourceType: resourceType,
		StartTime:    startTime,
		EndTime:      endTime,
		Page:         page,
		PageSize:     pageSize,
	}

	// 调用应用服务层查询审计日志。
	list, total, err := h.service.QueryLogs(c.Request.Context(), query)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to query audit logs", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to query audit logs", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Audit logs queried successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreatePolicy 处理创建审计策略的HTTP请求。
// Method: POST
// Path: /audit/policies
func (h *Handler) CreatePolicy(c *gin.Context) {
	// 定义请求体结构，用于接收策略名称和描述。
	var req struct {
		Name        string `json:"name" binding:"required"` // 策略名称，必填。
		Description string `json:"description"`             // 描述，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建审计策略。
	policy, err := h.service.CreatePolicy(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create audit policy", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create audit policy", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Audit policy created successfully", policy)
}

// ListPolicies 处理列出审计策略的HTTP请求。
// Method: GET
// Path: /audit/policies
func (h *Handler) ListPolicies(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取审计策略列表。
	list, total, err := h.service.ListPolicies(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list audit policies", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list audit policies", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Audit policies listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateReport 处理创建审计报告的HTTP请求。
// Method: POST
// Path: /audit/reports
func (h *Handler) CreateReport(c *gin.Context) {
	// 定义请求体结构，用于接收报告标题和描述。
	var req struct {
		Title       string `json:"title" binding:"required"` // 报告标题，必填。
		Description string `json:"description"`              // 描述，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建审计报告。
	report, err := h.service.CreateReport(c.Request.Context(), req.Title, req.Description)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create audit report", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create audit report", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Audit report created successfully", report)
}

// ListReports 处理列出审计报告的HTTP请求。
// Method: GET
// Path: /audit/reports
func (h *Handler) ListReports(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取审计报告列表。
	list, total, err := h.service.ListReports(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list audit reports", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list audit reports", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Audit reports listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册Audit模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /audit 路由组，用于所有审计相关接口。
	group := r.Group("/audit")
	{
		group.GET("/logs", h.QueryLogs) // 查询审计日志。

		group.POST("/policies", h.CreatePolicy) // 创建审计策略。
		group.GET("/policies", h.ListPolicies)  // 获取审计策略列表。
		// TODO: 补充更新、删除审计策略的接口。

		group.POST("/reports", h.CreateReport) // 创建审计报告。
		group.GET("/reports", h.ListReports)   // 获取审计报告列表。
		// TODO: 补充生成、更新、删除审计报告的接口。
	}
}

package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/data_ingestion/application"   // 导入数据摄取模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/data_ingestion/domain/entity" // 导入数据摄取模块的领域实体。
	"github.com/wyfcoding/pkg/response"                                    // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了DataIngestion模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.DataIngestionService // 依赖DataIngestion应用服务，处理核心业务逻辑。
	logger  *slog.Logger                      // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 DataIngestion HTTP Handler 实例。
func NewHandler(service *application.DataIngestionService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterSource 处理注册数据源的HTTP请求。
// Method: POST
// Path: /ingestion/sources
func (h *Handler) RegisterSource(c *gin.Context) {
	// 定义请求体结构，用于接收数据源的注册信息。
	var req struct {
		Name        string            `json:"name" binding:"required"`   // 数据源名称，必填。
		Type        entity.SourceType `json:"type" binding:"required"`   // 数据源类型，必填。
		Config      string            `json:"config" binding:"required"` // 连接配置，必填（JSON字符串）。
		Description string            `json:"description"`               // 描述，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层注册数据源。
	source, err := h.service.RegisterSource(c.Request.Context(), req.Name, req.Type, req.Config, req.Description)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to register source", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to register source", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Source registered successfully", source)
}

// ListSources 处理获取数据源列表的HTTP请求。
// Method: GET
// Path: /ingestion/sources
func (h *Handler) ListSources(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取数据源列表。
	list, total, err := h.service.ListSources(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list sources", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list sources", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Sources listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// TriggerIngestion 处理触发数据摄取任务的HTTP请求。
// Method: POST
// Path: /ingestion/sources/:id/trigger
func (h *Handler) TriggerIngestion(c *gin.Context) {
	// 从URL路径中解析数据源ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层触发数据摄取任务。
	job, err := h.service.TriggerIngestion(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to trigger ingestion", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to trigger ingestion", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Ingestion triggered successfully", job)
}

// ListJobs 处理获取数据摄取任务列表的HTTP请求。
// Method: GET
// Path: /ingestion/jobs
func (h *Handler) ListJobs(c *gin.Context) {
	// 从查询参数中获取数据源ID、页码和每页大小，并设置默认值。
	sourceID, _ := strconv.ParseUint(c.Query("source_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取数据摄取任务列表。
	list, total, err := h.service.ListJobs(c.Request.Context(), sourceID, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list jobs", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list jobs", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Jobs listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册DataIngestion模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /ingestion 路由组，用于所有数据摄取相关接口。
	group := r.Group("/ingestion")
	{
		group.POST("/sources", h.RegisterSource)               // 注册数据源。
		group.GET("/sources", h.ListSources)                   // 获取数据源列表。
		group.POST("/sources/:id/trigger", h.TriggerIngestion) // 触发数据摄取任务。
		group.GET("/jobs", h.ListJobs)                         // 获取数据摄取任务列表。
		// TODO: 补充获取数据源详情、更新数据源、删除数据源、获取任务详情等接口。
	}
}

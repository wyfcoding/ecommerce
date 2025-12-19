package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/scheduler/application" // 导入调度模块的应用服务。
	"github.com/wyfcoding/pkg/response"                             // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了Scheduler模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.SchedulerService // 依赖Scheduler应用服务，处理核心业务逻辑。
	logger  *slog.Logger                  // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Scheduler HTTP Handler 实例。
func NewHandler(service *application.SchedulerService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateJob 处理创建定时任务的HTTP请求。
// Method: POST
// Path: /scheduler/jobs
func (h *Handler) CreateJob(c *gin.Context) {
	// 定义请求体结构，用于接收定时任务的创建信息。
	var req struct {
		Name        string `json:"name" binding:"required"`      // 任务名称，必填。
		Description string `json:"description"`                  // 任务描述，选填。
		CronExpr    string `json:"cron_expr" binding:"required"` // Cron表达式，必填。
		Handler     string `json:"handler" binding:"required"`   // 处理器名称，必填。
		Params      string `json:"params"`                       // 参数，选填（通常为JSON字符串）。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建任务。
	job, err := h.service.CreateJob(c.Request.Context(), req.Name, req.Description, req.CronExpr, req.Handler, req.Params)
	if err != nil {
		h.logger.Error("Failed to create job", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create job", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Job created successfully", job)
}

// UpdateJob 处理更新定时任务的HTTP请求。
// Method: PUT
// Path: /scheduler/jobs/:id
func (h *Handler) UpdateJob(c *gin.Context) {
	// 从URL路径中解析任务ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Job ID", err.Error())
		return
	}

	// 定义请求体结构，用于接收任务更新信息。
	var req struct {
		CronExpr string `json:"cron_expr" binding:"required"` // 新的Cron表达式，必填。
		Params   string `json:"params"`                       // 新的参数，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层更新任务。
	if err := h.service.UpdateJob(c.Request.Context(), id, req.CronExpr, req.Params); err != nil {
		h.logger.Error("Failed to update job", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update job", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Job updated successfully", nil)
}

// ToggleJobStatus 处理切换定时任务状态的HTTP请求（启用或禁用）。
// Method: PUT
// Path: /scheduler/jobs/:id/status
func (h *Handler) ToggleJobStatus(c *gin.Context) {
	// 从URL路径中解析任务ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Job ID", err.Error())
		return
	}

	// 定义请求体结构，用于接收启用/禁用标志。
	var req struct {
		Enable bool `json:"enable"` // 启用或禁用，必填。
	}
	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层切换任务状态。
	if err := h.service.ToggleJobStatus(c.Request.Context(), id, req.Enable); err != nil {
		h.logger.Error("Failed to toggle job status", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to toggle job status", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Job status updated successfully", nil)
}

// RunJob 处理立即运行定时任务的HTTP请求。
// Method: POST
// Path: /scheduler/jobs/:id/run
func (h *Handler) RunJob(c *gin.Context) {
	// 从URL路径中解析任务ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Job ID", err.Error())
		return
	}

	// 调用应用服务层立即运行任务。
	if err := h.service.RunJob(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to run job", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to run job", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Job started successfully", nil)
}

// ListJobs 处理获取定时任务列表的HTTP请求。
// Method: GET
// Path: /scheduler/jobs
func (h *Handler) ListJobs(c *gin.Context) {
	// 从查询参数中获取状态字符串、页码和每页大小，并设置默认值。
	statusStr := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var status *int
	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil { // 只有当状态字符串能成功转换为int时才设置过滤状态。
			status = &s
		}
	}

	// 调用应用服务层获取任务列表。
	list, total, err := h.service.ListJobs(c.Request.Context(), status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list jobs", "error", err)
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

// ListJobLogs 处理获取任务日志列表的HTTP请求。
// Method: GET
// Path: /scheduler/logs
func (h *Handler) ListJobLogs(c *gin.Context) {
	// 从查询参数中获取任务ID、页码和每页大小，并设置默认值。
	jobID, _ := strconv.ParseUint(c.Query("job_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取任务日志列表。
	list, total, err := h.service.ListJobLogs(c.Request.Context(), jobID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list job logs", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list job logs", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Job logs listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册Scheduler模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /scheduler 路由组，用于所有调度任务相关接口。
	group := r.Group("/scheduler")
	{
		group.POST("/jobs", h.CreateJob)                 // 创建任务。
		group.PUT("/jobs/:id", h.UpdateJob)              // 更新任务。
		group.PUT("/jobs/:id/status", h.ToggleJobStatus) // 切换任务状态。
		group.POST("/jobs/:id/run", h.RunJob)            // 立即运行任务。
		group.GET("/jobs", h.ListJobs)                   // 获取任务列表。
		group.GET("/logs", h.ListJobLogs)                // 获取任务日志列表。
		// TODO: 补充获取任务详情、获取任务日志详情等接口。
	}
}

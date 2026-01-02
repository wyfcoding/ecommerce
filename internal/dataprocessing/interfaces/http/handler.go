package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/dataprocessing/application" // 导入数据处理模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/dataprocessing/domain"      // 导入数据处理模块的领域实体。
	"github.com/wyfcoding/pkg/response"                                  // 导入统一的响应处理工具。

	"log/slog" // 导入结构化日志库。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
)

// Handler 结构体定义了DataProcessing模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.DataProcessingService // 依赖DataProcessing应用服务，处理核心业务逻辑。
	logger  *slog.Logger                       // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 DataProcessing HTTP Handler 实例。
func NewHandler(service *application.DataProcessingService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// SubmitTask 处理提交数据处理任务的HTTP请求。
// HTTP 方法: POST
// 请求路径: /processing/tasks
func (h *Handler) SubmitTask(c *gin.Context) {
	// 定义请求体结构，用于接收任务的提交信息。
	var req struct {
		Name       string `json:"name" binding:"required"` // 任务名称，必填。
		Type       string `json:"type" binding:"required"` // 任务类型，必填。
		Config     string `json:"config"`                  // 配置，选填（JSON字符串）。
		WorkflowID uint64 `json:"workflow_id"`             // 所属工作流ID，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层提交任务。
	task, err := h.service.SubmitTask(c.Request.Context(), req.Name, req.Type, req.Config, req.WorkflowID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to submit task", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to submit task", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Task submitted successfully", task)
}

// GetTask 处理获取任务详情的HTTP请求。
func (h *Handler) GetTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid task ID", err.Error())
		return
	}

	task, err := h.service.GetTask(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get task", "id", id, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get task", err.Error())
		return
	}
	if task == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "Task not found", "")
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Task retrieved successfully", task)
}

// CancelTask 处理取消任务的HTTP请求。
func (h *Handler) CancelTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid task ID", err.Error())
		return
	}

	if err := h.service.CancelTask(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to cancel task", "id", id, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to cancel task", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Task cancelled successfully", nil)
}

// ListTasks 处理获取数据处理任务列表的HTTP请求。
func (h *Handler) ListTasks(c *gin.Context) {
	// 从查询参数中获取工作流ID、任务状态、页码和每页大小，并设置默认值。
	var (
		workflowID uint64
		err        error
	)
	if val := c.Query("workflow_id"); val != "" {
		workflowID, err = strconv.ParseUint(val, 10, 64)
		if err != nil {
			response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid workflow_id", err.Error())
			return
		}
	}

	var status int
	if val := c.Query("status"); val != "" {
		status, err = strconv.Atoi(val)
		if err != nil {
			response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid status", err.Error())
			return
		}
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	// 调用应用服务层获取任务列表。
	list, total, err := h.service.ListTasks(c.Request.Context(), workflowID, domain.TaskStatus(status), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list tasks", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list tasks", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Tasks listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateWorkflow 处理创建工作流的HTTP请求。
// HTTP 方法: POST
// 请求路径: /processing/workflows
func (h *Handler) CreateWorkflow(c *gin.Context) {
	// 定义请求体结构，用于接收工作流的创建信息。
	var req struct {
		Name        string `json:"name" binding:"required"`  // 工作流名称，必填。
		Description string `json:"description"`              // 描述，选填。
		Steps       string `json:"steps" binding:"required"` // 步骤定义，必填（JSON字符串）。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建工作流。
	workflow, err := h.service.CreateWorkflow(c.Request.Context(), req.Name, req.Description, req.Steps)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create workflow", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create workflow", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Workflow created successfully", workflow)
}

// GetWorkflow 处理获取工作流详情的HTTP请求。
func (h *Handler) GetWorkflow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid workflow ID", err.Error())
		return
	}

	workflow, err := h.service.GetWorkflow(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get workflow", "id", id, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get workflow", err.Error())
		return
	}
	if workflow == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "Workflow not found", "")
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Workflow retrieved successfully", workflow)
}

// UpdateWorkflow 处理更新工作流的HTTP请求。
func (h *Handler) UpdateWorkflow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid workflow ID", err.Error())
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Steps       string `json:"steps"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.UpdateWorkflow(c.Request.Context(), id, req.Name, req.Description, req.Steps); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to update workflow", "id", id, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update workflow", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Workflow updated successfully", nil)
}

// SetWorkflowActive 处理激活/停用工作流的HTTP请求。
func (h *Handler) SetWorkflowActive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid workflow ID", err.Error())
		return
	}

	var req struct {
		Active bool `json:"active" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.SetWorkflowActive(c.Request.Context(), id, req.Active); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to set workflow status", "id", id, "active", req.Active, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to set workflow status", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Workflow status updated successfully", nil)
}

// ListWorkflows 处理获取工作流列表的HTTP请求。
// HTTP 方法: GET
// 请求路径: /processing/workflows
func (h *Handler) ListWorkflows(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	// 调用应用服务层获取工作流列表。
	list, total, err := h.service.ListWorkflows(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list workflows", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list workflows", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Workflows listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册DataProcessing模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /processing 路由组，用于所有数据处理相关接口。
	group := r.Group("/processing")
	{
		// 任务管理接口。
		group.POST("/tasks", h.SubmitTask)            // 提交任务。
		group.GET("/tasks", h.ListTasks)              // 获取任务列表。
		group.GET("/tasks/:id", h.GetTask)            // 获取任务详情。
		group.POST("/tasks/:id/cancel", h.CancelTask) // 取消任务。

		// 工作流管理接口。
		group.POST("/workflows", h.CreateWorkflow)              // 创建工作流。
		group.GET("/workflows", h.ListWorkflows)                // 获取工作流列表。
		group.GET("/workflows/:id", h.GetWorkflow)              // 获取工作流详情。
		group.PUT("/workflows/:id", h.UpdateWorkflow)           // 更新工作流。
		group.PUT("/workflows/:id/active", h.SetWorkflowActive) // 激活/停用工作流。
	}
}

package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/data_processing/application"   // 导入数据处理模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/data_processing/domain/entity" // 导入数据处理模块的领域实体。
	"github.com/wyfcoding/pkg/response"                                     // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
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
// Method: POST
// Path: /processing/tasks
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

// ListTasks 处理获取数据处理任务列表的HTTP请求。
// Method: GET
// Path: /processing/tasks
func (h *Handler) ListTasks(c *gin.Context) {
	// 从查询参数中获取工作流ID、任务状态、页码和每页大小，并设置默认值。
	workflowID, _ := strconv.ParseUint(c.Query("workflow_id"), 10, 64)
	status, _ := strconv.Atoi(c.Query("status"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取任务列表。
	list, total, err := h.service.ListTasks(c.Request.Context(), workflowID, entity.TaskStatus(status), page, pageSize)
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
// Method: POST
// Path: /processing/workflows
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

// ListWorkflows 处理获取工作流列表的HTTP请求。
// Method: GET
// Path: /processing/workflows
func (h *Handler) ListWorkflows(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

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
		group.POST("/tasks", h.SubmitTask) // 提交任务。
		group.GET("/tasks", h.ListTasks)   // 获取任务列表。
		// TODO: 补充获取任务详情、更新任务、取消任务等接口。

		// 工作流管理接口。
		group.POST("/workflows", h.CreateWorkflow) // 创建工作流。
		group.GET("/workflows", h.ListWorkflows)   // 获取工作流列表。
		// TODO: 补充获取工作流详情、更新工作流、激活/停用工作流等接口。
	}
}

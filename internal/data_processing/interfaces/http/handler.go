package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/data_processing/application"
	"github.com/wyfcoding/ecommerce/internal/data_processing/domain/entity"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"log/slog"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *application.DataProcessingService
	logger  *slog.Logger
}

func NewHandler(service *application.DataProcessingService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// SubmitTask 提交任务
func (h *Handler) SubmitTask(c *gin.Context) {
	var req struct {
		Name       string `json:"name" binding:"required"`
		Type       string `json:"type" binding:"required"`
		Config     string `json:"config"`
		WorkflowID uint64 `json:"workflow_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	task, err := h.service.SubmitTask(c.Request.Context(), req.Name, req.Type, req.Config, req.WorkflowID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to submit task", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to submit task", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Task submitted successfully", task)
}

// ListTasks 获取任务列表
func (h *Handler) ListTasks(c *gin.Context) {
	workflowID, _ := strconv.ParseUint(c.Query("workflow_id"), 10, 64)
	status, _ := strconv.Atoi(c.Query("status"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListTasks(c.Request.Context(), workflowID, entity.TaskStatus(status), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list tasks", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list tasks", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Tasks listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateWorkflow 创建工作流
func (h *Handler) CreateWorkflow(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Steps       string `json:"steps" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	workflow, err := h.service.CreateWorkflow(c.Request.Context(), req.Name, req.Description, req.Steps)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create workflow", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create workflow", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Workflow created successfully", workflow)
}

// ListWorkflows 获取工作流列表
func (h *Handler) ListWorkflows(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListWorkflows(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list workflows", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list workflows", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Workflows listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/processing")
	{
		group.POST("/tasks", h.SubmitTask)
		group.GET("/tasks", h.ListTasks)
		group.POST("/workflows", h.CreateWorkflow)
		group.GET("/workflows", h.ListWorkflows)
	}
}

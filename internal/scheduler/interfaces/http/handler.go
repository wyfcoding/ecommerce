package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/scheduler/application"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 处理 HTTP 或 gRPC 请求。
type Handler struct {
	service *application.SchedulerService
	logger  *slog.Logger
}

// NewHandler 处理 HTTP 或 gRPC 请求。
func NewHandler(service *application.SchedulerService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) CreateJob(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		CronExpr    string `json:"cron_expr" binding:"required"`
		Handler     string `json:"handler" binding:"required"`
		Params      string `json:"params"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	job, err := h.service.CreateJob(c.Request.Context(), req.Name, req.Description, req.CronExpr, req.Handler, req.Params)
	if err != nil {
		h.logger.Error("Failed to create job", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create job", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Job created successfully", job)
}

func (h *Handler) UpdateJob(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Job ID", err.Error())
		return
	}

	var req struct {
		CronExpr string `json:"cron_expr" binding:"required"`
		Params   string `json:"params"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.UpdateJob(c.Request.Context(), id, req.CronExpr, req.Params); err != nil {
		h.logger.Error("Failed to update job", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update job", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Job updated successfully", nil)
}

func (h *Handler) ToggleJobStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Job ID", err.Error())
		return
	}

	var req struct {
		Enable bool `json:"enable"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.ToggleJobStatus(c.Request.Context(), id, req.Enable); err != nil {
		h.logger.Error("Failed to toggle job status", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to toggle job status", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Job status updated successfully", nil)
}

func (h *Handler) RunJob(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Job ID", err.Error())
		return
	}

	if err := h.service.RunJob(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to run job", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to run job", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Job started successfully", nil)
}

func (h *Handler) ListJobs(c *gin.Context) {
	statusStr := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var status *int
	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil {
			status = &s
		}
	}

	list, total, err := h.service.ListJobs(c.Request.Context(), status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list jobs", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list jobs", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Jobs listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) ListJobLogs(c *gin.Context) {
	jobID, _ := strconv.ParseUint(c.Query("job_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListJobLogs(c.Request.Context(), jobID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list job logs", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list job logs", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Job logs listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/scheduler")
	{
		group.POST("/jobs", h.CreateJob)
		group.PUT("/jobs/:id", h.UpdateJob)
		group.PUT("/jobs/:id/status", h.ToggleJobStatus)
		group.POST("/jobs/:id/run", h.RunJob)
		group.GET("/jobs", h.ListJobs)
		group.GET("/logs", h.ListJobLogs)
	}
}

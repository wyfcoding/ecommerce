package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/dataingestion/application"
	"github.com/wyfcoding/ecommerce/internal/dataingestion/domain"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 结构体定义了DataIngestion模块的HTTP处理层。
type Handler struct {
	app    *application.DataIngestionService
	logger *slog.Logger
}

// NewHandler 创建并返回一个新的 DataIngestion HTTP Handler 实例。
func NewHandler(app *application.DataIngestionService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// RegisterSource 处理注册数据源的HTTP请求。
func (h *Handler) RegisterSource(c *gin.Context) {
	var req struct {
		Name        string            `json:"name" binding:"required"`
		Type        domain.SourceType `json:"type" binding:"required"`
		Config      string            `json:"config" binding:"required"`
		Description string            `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	source, err := h.app.RegisterSource(c.Request.Context(), req.Name, req.Type, req.Config, req.Description)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to register source", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to register source", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Source registered successfully", source)
}

// ListSources 处理获取数据源列表的HTTP请求。
func (h *Handler) ListSources(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	list, total, err := h.app.ListSources(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list sources", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list sources", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Sources listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// TriggerIngestion 处理触发数据摄取任务的HTTP请求。
func (h *Handler) TriggerIngestion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	job, err := h.app.TriggerIngestion(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to trigger ingestion", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to trigger ingestion", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Ingestion triggered successfully", job)
}

// ListJobs 处理获取数据摄取任务列表的HTTP请求。
func (h *Handler) ListJobs(c *gin.Context) {
	var (
		sourceID uint64
		err      error
	)
	if val := c.Query("source_id"); val != "" {
		sourceID, err = strconv.ParseUint(val, 10, 64)
		if err != nil {
			response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid source_id", err.Error())
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

	list, total, err := h.app.ListJobs(c.Request.Context(), sourceID, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list jobs", "error", err)
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

// RegisterRoutes 在给定的Gin路由组中注册DataIngestion模块的HTTP路由。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/ingestion")
	{
		group.POST("/sources", h.RegisterSource)
		group.GET("/sources", h.ListSources)
		group.POST("/sources/:id/trigger", h.TriggerIngestion)
		group.GET("/jobs", h.ListJobs)
	}
}

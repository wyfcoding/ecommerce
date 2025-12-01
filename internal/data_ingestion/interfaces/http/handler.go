package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/data_ingestion/application"
	"github.com/wyfcoding/ecommerce/internal/data_ingestion/domain/entity"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"log/slog"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *application.DataIngestionService
	logger  *slog.Logger
}

func NewHandler(service *application.DataIngestionService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterSource 注册数据源
func (h *Handler) RegisterSource(c *gin.Context) {
	var req struct {
		Name        string            `json:"name" binding:"required"`
		Type        entity.SourceType `json:"type" binding:"required"`
		Config      string            `json:"config" binding:"required"`
		Description string            `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	source, err := h.service.RegisterSource(c.Request.Context(), req.Name, req.Type, req.Config, req.Description)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to register source", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to register source", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Source registered successfully", source)
}

// ListSources 获取数据源列表
func (h *Handler) ListSources(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListSources(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list sources", "error", err)
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

// TriggerIngestion 触发数据摄取
func (h *Handler) TriggerIngestion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	job, err := h.service.TriggerIngestion(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to trigger ingestion", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to trigger ingestion", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Ingestion triggered successfully", job)
}

// ListJobs 获取任务列表
func (h *Handler) ListJobs(c *gin.Context) {
	sourceID, _ := strconv.ParseUint(c.Query("source_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListJobs(c.Request.Context(), sourceID, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list jobs", "error", err)
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

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/ingestion")
	{
		group.POST("/sources", h.RegisterSource)
		group.GET("/sources", h.ListSources)
		group.POST("/sources/:id/trigger", h.TriggerIngestion)
		group.GET("/jobs", h.ListJobs)
	}
}

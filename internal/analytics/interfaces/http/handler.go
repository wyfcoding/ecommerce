package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/analytics/application"
	"github.com/wyfcoding/ecommerce/internal/analytics/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/analytics/domain/repository"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.AnalyticsService
	logger  *slog.Logger
}

func NewHandler(service *application.AnalyticsService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RecordMetric 记录指标
func (h *Handler) RecordMetric(c *gin.Context) {
	var req struct {
		MetricType   string  `json:"metric_type" binding:"required"`
		Name         string  `json:"name" binding:"required"`
		Value        float64 `json:"value" binding:"required"`
		Granularity  string  `json:"granularity" binding:"required"`
		Dimension    string  `json:"dimension"`
		DimensionVal string  `json:"dimension_val"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	err := h.service.RecordMetric(c.Request.Context(), entity.MetricType(req.MetricType), req.Name, req.Value, entity.TimeGranularity(req.Granularity), req.Dimension, req.DimensionVal)
	if err != nil {
		h.logger.Error("Failed to record metric", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to record metric", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Metric recorded successfully", nil)
}

// QueryMetrics 查询指标
func (h *Handler) QueryMetrics(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	metricType := c.DefaultQuery("metric_type", "")
	granularity := c.DefaultQuery("granularity", "")
	startTimeStr := c.DefaultQuery("start_time", "")
	endTimeStr := c.DefaultQuery("end_time", "")

	var startTime, endTime time.Time
	if startTimeStr != "" {
		startTime, _ = time.Parse(time.RFC3339, startTimeStr)
	}
	if endTimeStr != "" {
		endTime, _ = time.Parse(time.RFC3339, endTimeStr)
	}

	query := &repository.MetricQuery{
		MetricType:  entity.MetricType(metricType),
		Granularity: entity.TimeGranularity(granularity),
		StartTime:   startTime,
		EndTime:     endTime,
		Page:        page,
		PageSize:    pageSize,
	}

	list, total, err := h.service.QueryMetrics(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to query metrics", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to query metrics", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Metrics queried successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateDashboard 创建仪表板
func (h *Handler) CreateDashboard(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		UserID      uint64 `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	dashboard, err := h.service.CreateDashboard(c.Request.Context(), req.Name, req.Description, req.UserID)
	if err != nil {
		h.logger.Error("Failed to create dashboard", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create dashboard", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Dashboard created successfully", dashboard)
}

// GetDashboard 获取仪表板详情
func (h *Handler) GetDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	dashboard, err := h.service.GetDashboard(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get dashboard", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get dashboard", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Dashboard details retrieved successfully", dashboard)
}

// AddMetricToDashboard 添加指标到仪表板
func (h *Handler) AddMetricToDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		MetricType string `json:"metric_type" binding:"required"`
		Title      string `json:"title" binding:"required"`
		ChartType  string `json:"chart_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	err = h.service.AddMetricToDashboard(c.Request.Context(), id, entity.MetricType(req.MetricType), req.Title, req.ChartType)
	if err != nil {
		h.logger.Error("Failed to add metric to dashboard", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add metric to dashboard", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Metric added to dashboard successfully", nil)
}

// CreateReport 创建报告
func (h *Handler) CreateReport(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		UserID      uint64 `json:"user_id" binding:"required"`
		ReportType  string `json:"report_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	report, err := h.service.CreateReport(c.Request.Context(), req.Title, req.Description, req.UserID, req.ReportType)
	if err != nil {
		h.logger.Error("Failed to create report", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create report", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Report created successfully", report)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/analytics")
	{
		group.POST("/metrics", h.RecordMetric)
		group.GET("/metrics", h.QueryMetrics)

		group.POST("/dashboards", h.CreateDashboard)
		group.GET("/dashboards/:id", h.GetDashboard)
		group.POST("/dashboards/:id/metrics", h.AddMetricToDashboard)

		group.POST("/reports", h.CreateReport)
	}
}

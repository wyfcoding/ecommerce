package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/analytics/application"
	"github.com/wyfcoding/ecommerce/internal/analytics/domain"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	app    *application.Analytics
	logger *slog.Logger
}

func NewHandler(app *application.Analytics, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

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
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid data", "")
		return
	}

	err := h.app.RecordMetric(c.Request.Context(), domain.MetricType(req.MetricType), req.Name, req.Value, domain.TimeGranularity(req.Granularity), req.Dimension, req.DimensionVal)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to record metric", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Created", nil)
}

func (h *Handler) QueryMetrics(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid page", "")
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid page_size", "")
		return
	}
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

	query := &domain.MetricQuery{
		MetricType:  domain.MetricType(metricType),
		Granularity: domain.TimeGranularity(granularity),
		StartTime:   startTime,
		EndTime:     endTime,
		Page:        page,
		PageSize:    pageSize,
	}

	list, total, err := h.app.QueryMetrics(c.Request.Context(), query)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to query metrics", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (h *Handler) CreateDashboard(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		UserID      uint64 `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid data", "")
		return
	}

	dashboard, err := h.app.CreateDashboard(c.Request.Context(), req.Name, req.Description, req.UserID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create dashboard", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Created", dashboard)
}

func (h *Handler) GetDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid ID", "")
		return
	}

	dashboard, err := h.app.GetDashboard(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get dashboard", "id", id, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, dashboard)
}

func (h *Handler) AddMetricToDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid ID", "")
		return
	}

	var req struct {
		MetricType string `json:"metric_type" binding:"required"`
		Title      string `json:"title" binding:"required"`
		ChartType  string `json:"chart_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid data", "")
		return
	}

	err = h.app.AddMetricToDashboard(c.Request.Context(), id, domain.MetricType(req.MetricType), req.Title, req.ChartType)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to add metric to dashboard", "id", id, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, nil)
}

func (h *Handler) UpdateDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid ID", "")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid data", "")
		return
	}

	dashboard, err := h.app.UpdateDashboard(c.Request.Context(), id, req.Name, req.Description)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, dashboard)
}

func (h *Handler) DeleteDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid ID", "")
		return
	}

	if err := h.app.DeleteDashboard(c.Request.Context(), id); err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, nil)
}

func (h *Handler) ListDashboards(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user_id", "")
		return
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid page", "")
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid page_size", "")
		return
	}

	dashboards, total, err := h.app.ListDashboards(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, gin.H{
		"list":  dashboards,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (h *Handler) PublishDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid ID", "")
		return
	}

	if err := h.app.PublishDashboard(c.Request.Context(), id); err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, nil)
}

func (h *Handler) CreateReport(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		UserID      uint64 `json:"user_id" binding:"required"`
		ReportType  string `json:"report_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid data", "")
		return
	}

	report, err := h.app.CreateReport(c.Request.Context(), req.Title, req.Description, req.UserID, req.ReportType)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create report", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Created", report)
}

func (h *Handler) GetReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid ID", "")
		return
	}

	report, err := h.app.GetReport(c.Request.Context(), id)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}
	if report == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "report not found", "")
		return
	}

	response.Success(c, report)
}

func (h *Handler) ListReports(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user_id", "")
		return
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid page", "")
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid page_size", "")
		return
	}

	reports, total, err := h.app.ListReports(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, gin.H{
		"list":  reports,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/analytics")
	{
		group.POST("/metrics", h.RecordMetric)
		group.GET("/metrics", h.QueryMetrics)
		group.POST("/dashboards", h.CreateDashboard)
		group.GET("/dashboards", h.ListDashboards)
		group.GET("/dashboards/:id", h.GetDashboard)
		group.PUT("/dashboards/:id", h.UpdateDashboard)
		group.DELETE("/dashboards/:id", h.DeleteDashboard)
		group.POST("/dashboards/:id/metrics", h.AddMetricToDashboard)
		group.POST("/dashboards/:id/publish", h.PublishDashboard)
		group.POST("/reports", h.CreateReport)
		group.GET("/reports", h.ListReports)
		group.GET("/reports/:id", h.GetReport)
	}
}

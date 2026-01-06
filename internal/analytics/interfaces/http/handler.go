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

// RecordMetric 处理记录单个业务指标的 HTTP 请求。
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
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	err := h.app.RecordMetric(c.Request.Context(), domain.MetricType(req.MetricType), req.Name, req.Value, domain.TimeGranularity(req.Granularity), req.Dimension, req.DimensionVal)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to record metric", "metric_type", req.MetricType, "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "metric recorded", nil)
}

// QueryMetrics 根据时间范围和维度查询聚合指标数据。
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
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			response.ErrorWithStatus(c, http.StatusBadRequest, "invalid start_time format", err.Error())
			return
		}
	}
	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			response.ErrorWithStatus(c, http.StatusBadRequest, "invalid end_time format", err.Error())
			return
		}
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
		h.logger.ErrorContext(c.Request.Context(), "failed to query metrics", "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, list, total, int32(page), int32(pageSize))
}

// CreateDashboard 创建一个新的个性化数据仪表板。
func (h *Handler) CreateDashboard(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		UserID      uint64 `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	dashboard, err := h.app.CreateDashboard(c.Request.Context(), req.Name, req.Description, req.UserID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to create dashboard", "user_id", req.UserID, "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "dashboard created", dashboard)
}

// GetDashboard 获取指定的仪表板详情。
func (h *Handler) GetDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id format", "")
		return
	}

	dashboard, err := h.app.GetDashboard(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get dashboard", "dashboard_id", id, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, dashboard)
}

// AddMetricToDashboard 将业务指标关联至特定仪表板。
func (h *Handler) AddMetricToDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id format", "")
		return
	}

	var req struct {
		MetricType string `json:"metric_type" binding:"required"`
		Title      string `json:"title" binding:"required"`
		ChartType  string `json:"chart_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	err = h.app.AddMetricToDashboard(c.Request.Context(), id, domain.MetricType(req.MetricType), req.Title, req.ChartType)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to add metric to dashboard", "dashboard_id", id, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// UpdateDashboard 修改仪表板基础信息。
func (h *Handler) UpdateDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id format", "")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	dashboard, err := h.app.UpdateDashboard(c.Request.Context(), id, req.Name, req.Description)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to update dashboard", "dashboard_id", id, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, dashboard)
}

// DeleteDashboard 物理删除指定仪表板。
func (h *Handler) DeleteDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id format", "")
		return
	}

	if err := h.app.DeleteDashboard(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to delete dashboard", "dashboard_id", id, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// ListDashboards 分页列出指定用户拥有的仪表板。
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
		h.logger.ErrorContext(c.Request.Context(), "failed to list dashboards", "user_id", userID, "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, dashboards, total, int32(page), int32(pageSize))
}

// PublishDashboard 将仪表板状态设为公开。
func (h *Handler) PublishDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id format", "")
		return
	}

	if err := h.app.PublishDashboard(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to publish dashboard", "dashboard_id", id, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// CreateReport 处理生成数据分析报告的请求。
func (h *Handler) CreateReport(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		UserID      uint64 `json:"user_id" binding:"required"`
		ReportType  string `json:"report_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	report, err := h.app.CreateReport(c.Request.Context(), req.Title, req.Description, req.UserID, req.ReportType)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to create report", "user_id", req.UserID, "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "report created", report)
}

// GetReport 获取单份报告的详细分析内容。
func (h *Handler) GetReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id format", "")
		return
	}

	report, err := h.app.GetReport(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get report", "report_id", id, "error", err)
		response.Error(c, err)
		return
	}
	if report == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "report not found", "")
		return
	}

	response.Success(c, report)
}

// ListReports 获取指定用户的历史报告列表。
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
		h.logger.ErrorContext(c.Request.Context(), "failed to list reports", "user_id", userID, "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, reports, total, int32(page), int32(pageSize))
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

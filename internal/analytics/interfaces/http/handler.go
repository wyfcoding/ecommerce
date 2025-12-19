package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。
	"time"     // 导入时间包，用于时间解析。

	"github.com/wyfcoding/ecommerce/internal/analytics/application" // 导入分析模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/analytics/domain"      // 导入分析模块的领域层。
	"github.com/wyfcoding/pkg/response"                             // 导入统一的响应处理工具。

	"log/slog" // 导入结构化日志库。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
)

// Handler 结构体定义了Analytics模块的HTTP处理层。
type Handler struct {
	app    *application.AnalyticsService // 依赖Analytics应用服务 facade。
	logger *slog.Logger                  // 日志记录器。
}

// NewHandler 创建并返回一个新的 Analytics HTTP Handler 实例。
func NewHandler(app *application.AnalyticsService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// RecordMetric 处理记录指标的HTTP请求。
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

	err := h.app.RecordMetric(c.Request.Context(), domain.MetricType(req.MetricType), req.Name, req.Value, domain.TimeGranularity(req.Granularity), req.Dimension, req.DimensionVal)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to record metric", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to record metric", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Metric recorded successfully", nil)
}

// QueryMetrics 处理查询指标的HTTP请求。
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

// CreateDashboard 处理创建仪表板的HTTP请求。
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

	dashboard, err := h.app.CreateDashboard(c.Request.Context(), req.Name, req.Description, req.UserID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create dashboard", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create dashboard", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Dashboard created successfully", dashboard)
}

// GetDashboard 处理获取仪表板详情的HTTP请求。
func (h *Handler) GetDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	dashboard, err := h.app.GetDashboard(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get dashboard", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get dashboard", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Dashboard details retrieved successfully", dashboard)
}

// AddMetricToDashboard 处理添加指标到仪表板的HTTP请求。
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

	err = h.app.AddMetricToDashboard(c.Request.Context(), id, domain.MetricType(req.MetricType), req.Title, req.ChartType)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to add metric to dashboard", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add metric to dashboard", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Metric added to dashboard successfully", nil)
}

// UpdateDashboard 处理更新仪表板的HTTP请求。
func (h *Handler) UpdateDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	dashboard, err := h.app.UpdateDashboard(c.Request.Context(), id, req.Name, req.Description)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update dashboard", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Dashboard updated successfully", dashboard)
}

// DeleteDashboard 处理删除仪表板的HTTP请求。
func (h *Handler) DeleteDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.app.DeleteDashboard(c.Request.Context(), id); err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete dashboard", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Dashboard deleted successfully", nil)
}

// ListDashboards 处理列出仪表板的HTTP请求。
func (h *Handler) ListDashboards(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	dashboards, total, err := h.app.ListDashboards(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list dashboards", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Dashboards listed successfully", gin.H{
		"data":      dashboards,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// PublishDashboard 处理发布仪表板的HTTP请求。
func (h *Handler) PublishDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.app.PublishDashboard(c.Request.Context(), id); err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to publish dashboard", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Dashboard published successfully", nil)
}

// CreateReport 处理创建报告的HTTP请求。
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

	report, err := h.app.CreateReport(c.Request.Context(), req.Title, req.Description, req.UserID, req.ReportType)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create report", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create report", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Report created successfully", report)
}

// GetReport 处理获取报告详情的HTTP请求。
func (h *Handler) GetReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	report, err := h.app.GetReport(c.Request.Context(), id)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get report", err.Error())
		return
	}
	if report == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "Report not found", "report not found")
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Report retrieved successfully", report)
}

// UpdateReport 处理更新报告的HTTP请求。
func (h *Handler) UpdateReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	report, err := h.app.UpdateReport(c.Request.Context(), id, req.Title, req.Description)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update report", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Report updated successfully", report)
}

// DeleteReport 处理删除报告的HTTP请求。
func (h *Handler) DeleteReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.app.DeleteReport(c.Request.Context(), id); err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete report", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Report deleted successfully", nil)
}

// ListReports 处理列出报告的HTTP请求。
func (h *Handler) ListReports(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	reports, total, err := h.app.ListReports(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list reports", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Reports listed successfully", gin.H{
		"data":      reports,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// PublishReport 处理发布报告的HTTP请求。
func (h *Handler) PublishReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.app.PublishReport(c.Request.Context(), id); err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to publish report", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Report published successfully", nil)
}

// RegisterRoutes 注册Analytics模块的HTTP路由。
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
		group.PUT("/reports/:id", h.UpdateReport)
		group.DELETE("/reports/:id", h.DeleteReport)
		group.POST("/reports/:id/publish", h.PublishReport)
	}
}

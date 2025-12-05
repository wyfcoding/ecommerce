package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。
	"time"     // 导入时间包，用于时间解析。

	"github.com/wyfcoding/ecommerce/internal/analytics/application"       // 导入分析模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/analytics/domain/entity"     // 导入分析模块的领域实体。
	"github.com/wyfcoding/ecommerce/internal/analytics/domain/repository" // 导入分析模块的领域仓储查询对象。
	"github.com/wyfcoding/ecommerce/pkg/response"                         // 导入统一的响应处理工具。

	"log/slog" // 导入结构化日志库。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
)

// Handler 结构体定义了Analytics模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.AnalyticsService // 依赖Analytics应用服务，处理核心业务逻辑。
	logger  *slog.Logger                  // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Analytics HTTP Handler 实例。
func NewHandler(service *application.AnalyticsService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RecordMetric 处理记录指标的HTTP请求。
// Method: POST
// Path: /analytics/metrics
func (h *Handler) RecordMetric(c *gin.Context) {
	// 定义请求体结构，用于接收指标数据。
	var req struct {
		MetricType   string  `json:"metric_type" binding:"required"` // 指标类型，必填。
		Name         string  `json:"name" binding:"required"`        // 指标名称，必填。
		Value        float64 `json:"value" binding:"required"`       // 指标值，必填。
		Granularity  string  `json:"granularity" binding:"required"` // 时间粒度，必填。
		Dimension    string  `json:"dimension"`                      // 维度名称，选填。
		DimensionVal string  `json:"dimension_val"`                  // 维度值，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层记录指标。
	err := h.service.RecordMetric(c.Request.Context(), entity.MetricType(req.MetricType), req.Name, req.Value, entity.TimeGranularity(req.Granularity), req.Dimension, req.DimensionVal)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to record metric", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to record metric", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Metric recorded successfully", nil)
}

// QueryMetrics 处理查询指标的HTTP请求。
// Method: GET
// Path: /analytics/metrics
func (h *Handler) QueryMetrics(c *gin.Context) {
	// 从查询参数中获取分页和过滤条件，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	metricType := c.DefaultQuery("metric_type", "")
	granularity := c.DefaultQuery("granularity", "")
	startTimeStr := c.DefaultQuery("start_time", "")
	endTimeStr := c.DefaultQuery("end_time", "")

	var startTime, endTime time.Time
	// 解析起始时间字符串。
	if startTimeStr != "" {
		startTime, _ = time.Parse(time.RFC3339, startTimeStr)
	}
	// 解析结束时间字符串。
	if endTimeStr != "" {
		endTime, _ = time.Parse(time.RFC3339, endTimeStr)
	}

	// 构建查询对象。
	query := &repository.MetricQuery{
		MetricType:  entity.MetricType(metricType),
		Granularity: entity.TimeGranularity(granularity),
		StartTime:   startTime,
		EndTime:     endTime,
		Page:        page,
		PageSize:    pageSize,
	}

	// 调用应用服务层查询指标。
	list, total, err := h.service.QueryMetrics(c.Request.Context(), query)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to query metrics", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to query metrics", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Metrics queried successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateDashboard 处理创建仪表板的HTTP请求。
// Method: POST
// Path: /analytics/dashboards
func (h *Handler) CreateDashboard(c *gin.Context) {
	// 定义请求体结构，用于接收仪表板的创建信息。
	var req struct {
		Name        string `json:"name" binding:"required"`    // 仪表板名称，必填。
		Description string `json:"description"`                // 描述，选填。
		UserID      uint64 `json:"user_id" binding:"required"` // 用户ID，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建仪表板。
	dashboard, err := h.service.CreateDashboard(c.Request.Context(), req.Name, req.Description, req.UserID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create dashboard", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create dashboard", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Dashboard created successfully", dashboard)
}

// GetDashboard 处理获取仪表板详情的HTTP请求。
// Method: GET
// Path: /analytics/dashboards/:id
func (h *Handler) GetDashboard(c *gin.Context) {
	// 从URL路径中解析仪表板ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层获取仪表板详情。
	dashboard, err := h.service.GetDashboard(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get dashboard", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get dashboard", err.Error())
		return
	}

	// 返回成功的响应，包含仪表板详情。
	response.SuccessWithStatus(c, http.StatusOK, "Dashboard details retrieved successfully", dashboard)
}

// AddMetricToDashboard 处理添加指标到仪表板的HTTP请求。
// Method: POST
// Path: /analytics/dashboards/:id/metrics
func (h *Handler) AddMetricToDashboard(c *gin.Context) {
	// 从URL路径中解析仪表板ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 定义请求体结构，用于接收要添加的指标信息。
	var req struct {
		MetricType string `json:"metric_type" binding:"required"` // 指标类型，必填。
		Title      string `json:"title" binding:"required"`       // 指标图表标题，必填。
		ChartType  string `json:"chart_type" binding:"required"`  // 图表类型，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层添加指标到仪表板。
	err = h.service.AddMetricToDashboard(c.Request.Context(), id, entity.MetricType(req.MetricType), req.Title, req.ChartType)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to add metric to dashboard", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add metric to dashboard", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Metric added to dashboard successfully", nil)
}

// UpdateDashboard 处理更新仪表板的HTTP请求。
// Method: PUT
// Path: /analytics/dashboards/:id
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

	dashboard, err := h.service.UpdateDashboard(c.Request.Context(), id, req.Name, req.Description)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update dashboard", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Dashboard updated successfully", dashboard)
}

// DeleteDashboard 处理删除仪表板的HTTP请求。
// Method: DELETE
// Path: /analytics/dashboards/:id
func (h *Handler) DeleteDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.DeleteDashboard(c.Request.Context(), id); err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete dashboard", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Dashboard deleted successfully", nil)
}

// ListDashboards 处理列出仪表板的HTTP请求。
// Method: GET
// Path: /analytics/dashboards
func (h *Handler) ListDashboards(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	dashboards, total, err := h.service.ListDashboards(c.Request.Context(), userID, page, pageSize)
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
// Method: POST
// Path: /analytics/dashboards/:id/publish
func (h *Handler) PublishDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.PublishDashboard(c.Request.Context(), id); err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to publish dashboard", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Dashboard published successfully", nil)
}

// CreateReport 处理创建报告的HTTP请求。
// Method: POST
// Path: /analytics/reports
func (h *Handler) CreateReport(c *gin.Context) {
	// 定义请求体结构，用于接收报告的创建信息。
	var req struct {
		Title       string `json:"title" binding:"required"`       // 报告标题，必填。
		Description string `json:"description"`                    // 报告描述，选填。
		UserID      uint64 `json:"user_id" binding:"required"`     // 用户ID，必填。
		ReportType  string `json:"report_type" binding:"required"` // 报告类型，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建报告。
	report, err := h.service.CreateReport(c.Request.Context(), req.Title, req.Description, req.UserID, req.ReportType)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create report", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create report", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Report created successfully", report)
}

// GetReport 处理获取报告详情的HTTP请求。
// Method: GET
// Path: /analytics/reports/:id
func (h *Handler) GetReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	report, err := h.service.GetReport(c.Request.Context(), id)
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
// Method: PUT
// Path: /analytics/reports/:id
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

	report, err := h.service.UpdateReport(c.Request.Context(), id, req.Title, req.Description)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update report", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Report updated successfully", report)
}

// DeleteReport 处理删除报告的HTTP请求。
// Method: DELETE
// Path: /analytics/reports/:id
func (h *Handler) DeleteReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.DeleteReport(c.Request.Context(), id); err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete report", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Report deleted successfully", nil)
}

// ListReports 处理列出报告的HTTP请求。
// Method: GET
// Path: /analytics/reports
func (h *Handler) ListReports(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	reports, total, err := h.service.ListReports(c.Request.Context(), userID, page, pageSize)
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
// Method: POST
// Path: /analytics/reports/:id/publish
func (h *Handler) PublishReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.PublishReport(c.Request.Context(), id); err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to publish report", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Report published successfully", nil)
}

// RegisterRoutes 在给定的Gin路由组中注册Analytics模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /analytics 路由组，用于所有分析相关接口。
	group := r.Group("/analytics")
	{
		// Metrics 接口。
		group.POST("/metrics", h.RecordMetric) // 记录指标。
		group.GET("/metrics", h.QueryMetrics)  // 查询指标。

		// Dashboards 接口。
		group.POST("/dashboards", h.CreateDashboard)                  // 创建仪表板。
		group.GET("/dashboards", h.ListDashboards)                    // 列出仪表板。
		group.GET("/dashboards/:id", h.GetDashboard)                  // 获取仪表板详情。
		group.PUT("/dashboards/:id", h.UpdateDashboard)               // 更新仪表板。
		group.DELETE("/dashboards/:id", h.DeleteDashboard)            // 删除仪表板。
		group.POST("/dashboards/:id/metrics", h.AddMetricToDashboard) // 添加指标到仪表板。
		group.POST("/dashboards/:id/publish", h.PublishDashboard)     // 发布仪表板。

		// Reports 接口。
		group.POST("/reports", h.CreateReport)              // 创建报告。
		group.GET("/reports", h.ListReports)                // 列出报告。
		group.GET("/reports/:id", h.GetReport)              // 获取报告详情。
		group.PUT("/reports/:id", h.UpdateReport)           // 更新报告。
		group.DELETE("/reports/:id", h.DeleteReport)        // 删除报告。
		group.POST("/reports/:id/publish", h.PublishReport) // 发布报告。
	}
}

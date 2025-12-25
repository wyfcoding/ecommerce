package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/audit/application"
	"github.com/wyfcoding/ecommerce/internal/audit/domain"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 处理 HTTP 或 gRPC 请求。
type Handler struct {
	app    *application.Audit
	logger *slog.Logger
}

// NewHandler 创建并返回一个新的 Audit HTTP Handler 实例。
func NewHandler(app *application.Audit, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// QueryLogs 处理查询审计日志的HTTP请求。
func (h *Handler) QueryLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userID, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)
	eventType := c.DefaultQuery("event_type", "")
	module := c.DefaultQuery("module", "")
	resourceType := c.DefaultQuery("resource_type", "")
	startTimeStr := c.DefaultQuery("start_time", "")
	endTimeStr := c.DefaultQuery("end_time", "")

	var startTime, endTime time.Time
	if startTimeStr != "" {
		startTime, _ = time.Parse(time.RFC3339, startTimeStr)
	}
	if endTimeStr != "" {
		endTime, _ = time.Parse(time.RFC3339, endTimeStr)
	}

	query := &domain.AuditLogQuery{
		UserID:       userID,
		EventType:    domain.AuditEventType(eventType),
		Module:       module,
		ResourceType: resourceType,
		StartTime:    startTime,
		EndTime:      endTime,
		Page:         page,
		PageSize:     pageSize,
	}

	list, total, err := h.app.QueryLogs(c.Request.Context(), query)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to query audit logs", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to query audit logs", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Audit logs queried successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreatePolicy 处理创建审计策略的HTTP请求。
func (h *Handler) CreatePolicy(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	policy, err := h.app.CreatePolicy(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create audit policy", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create audit policy", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Audit policy created successfully", policy)
}

// ListPolicies 处理列出审计策略的HTTP请求。
func (h *Handler) ListPolicies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.app.ListPolicies(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list audit policies", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list audit policies", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Audit policies listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateReport 处理创建审计报告的HTTP请求。
func (h *Handler) CreateReport(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	report, err := h.app.CreateReport(c.Request.Context(), req.Title, req.Description)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create audit report", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create audit report", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Audit report created successfully", report)
}

// ListReports 处理列出审计报告的HTTP请求。
func (h *Handler) ListReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.app.ListReports(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list audit reports", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list audit reports", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Audit reports listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册Audit模块的HTTP路由。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/audit")
	{
		group.GET("/logs", h.QueryLogs)
		group.POST("/policies", h.CreatePolicy)
		group.GET("/policies", h.ListPolicies)
		group.PUT("/policies/:id", h.UpdatePolicy)
		group.POST("/reports", h.CreateReport)
		group.GET("/reports", h.ListReports)
		group.POST("/reports/:id/generate", h.GenerateReport)
	}
}

// UpdatePolicy 处理更新审计策略的HTTP请求。
func (h *Handler) UpdatePolicy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		EventTypes []string `json:"event_types"`
		Modules    []string `json:"modules"`
		Enabled    bool     `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.app.UpdatePolicy(c.Request.Context(), id, req.EventTypes, req.Modules, req.Enabled); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to update audit policy", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update audit policy", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Audit policy updated successfully", nil)
}

// GenerateReport 处理生成审计报告内容的HTTP请求。
func (h *Handler) GenerateReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.app.GenerateReport(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to generate audit report", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to generate audit report", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Audit report generated successfully", nil)
}

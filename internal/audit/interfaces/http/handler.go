package http

import (
	"net/http"
	"strconv"
	"time"

	"ecommerce/internal/audit/application"
	"ecommerce/internal/audit/domain/entity"
	"ecommerce/internal/audit/domain/repository"
	"ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.AuditService
	logger  *slog.Logger
}

func NewHandler(service *application.AuditService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// QueryLogs 查询审计日志
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

	query := &repository.AuditLogQuery{
		UserID:       userID,
		EventType:    entity.AuditEventType(eventType),
		Module:       module,
		ResourceType: resourceType,
		StartTime:    startTime,
		EndTime:      endTime,
		Page:         page,
		PageSize:     pageSize,
	}

	list, total, err := h.service.QueryLogs(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to query audit logs", "error", err)
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

// CreatePolicy 创建审计策略
func (h *Handler) CreatePolicy(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	policy, err := h.service.CreatePolicy(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		h.logger.Error("Failed to create audit policy", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create audit policy", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Audit policy created successfully", policy)
}

// ListPolicies 获取审计策略列表
func (h *Handler) ListPolicies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListPolicies(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list audit policies", "error", err)
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

// CreateReport 创建审计报告
func (h *Handler) CreateReport(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	report, err := h.service.CreateReport(c.Request.Context(), req.Title, req.Description)
	if err != nil {
		h.logger.Error("Failed to create audit report", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create audit report", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Audit report created successfully", report)
}

// ListReports 获取审计报告列表
func (h *Handler) ListReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListReports(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list audit reports", "error", err)
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

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/audit")
	{
		group.GET("/logs", h.QueryLogs)

		group.POST("/policies", h.CreatePolicy)
		group.GET("/policies", h.ListPolicies)

		group.POST("/reports", h.CreateReport)
		group.GET("/reports", h.ListReports)
	}
}

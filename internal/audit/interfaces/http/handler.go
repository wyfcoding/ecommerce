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

type Handler struct {
	app    *application.Audit
	logger *slog.Logger
}

func NewHandler(app *application.Audit, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

func (h *Handler) QueryLogs(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		response.BadRequest(c, "invalid page")
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil {
		response.BadRequest(c, "invalid page_size")
		return
	}
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
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (h *Handler) CreatePolicy(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	policy, err := h.app.CreatePolicy(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create audit policy", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Created", policy)
}

func (h *Handler) ListPolicies(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		response.BadRequest(c, "invalid page")
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		response.BadRequest(c, "invalid page_size")
		return
	}

	list, total, err := h.app.ListPolicies(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list audit policies", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (h *Handler) UpdatePolicy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid ID")
		return
	}

	var req struct {
		EventTypes []string `json:"event_types"`
		Modules    []string `json:"modules"`
		Enabled    bool     `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid data")
		return
	}

	if err := h.app.UpdatePolicy(c.Request.Context(), id, req.EventTypes, req.Modules, req.Enabled); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to update audit policy", "id", id, "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/audit")
	{
		group.GET("/logs", h.QueryLogs)
		group.POST("/policies", h.CreatePolicy)
		group.GET("/policies", h.ListPolicies)
		group.PUT("/policies/:id", h.UpdatePolicy)
	}
}

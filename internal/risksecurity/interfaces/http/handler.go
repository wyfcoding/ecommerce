package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/risksecurity/application"
	"github.com/wyfcoding/pkg/response"
	"github.com/wyfcoding/pkg/utils/ctxutil"

	"github.com/gin-gonic/gin"
)

// Handler 结构体定义了RiskSecurity模块的HTTP处理层。
type Handler struct {
	app    *application.RiskService
	logger *slog.Logger
}

// NewHandler 创建并返回一个新的 RiskSecurity HTTP Handler 实例。
func NewHandler(app *application.RiskService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// EvaluateRisk 处理评估风险的HTTP请求。
func (h *Handler) EvaluateRisk(c *gin.Context) {
	var req struct {
		UserID   uint64 `json:"user_id" binding:"required"`
		IP       string `json:"ip" binding:"required"`
		DeviceID string `json:"device_id"`
		Amount   int64  `json:"amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ctx := ctxutil.WithUserAgent(c.Request.Context(), c.Request.UserAgent())
	result, err := h.app.EvaluateRisk(ctx, req.UserID, req.IP, req.DeviceID, req.Amount)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to evaluate risk", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to evaluate risk", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Risk evaluated successfully", result)
}

// AddToBlacklist 处理添加实体到黑名单的HTTP请求。
func (h *Handler) AddToBlacklist(c *gin.Context) {
	var req struct {
		Type     string `json:"type" binding:"required"`
		Value    string `json:"value" binding:"required"`
		Reason   string `json:"reason"`
		Duration string `json:"duration"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		duration = 24 * time.Hour
	}

	if err := h.app.AddToBlacklist(c.Request.Context(), req.Type, req.Value, req.Reason, duration); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to add to blacklist", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add to blacklist", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Added to blacklist successfully", nil)
}

// RemoveFromBlacklist 处理从黑名单中移除实体的HTTP请求。
func (h *Handler) RemoveFromBlacklist(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.app.RemoveFromBlacklist(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to remove from blacklist", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to remove from blacklist", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Removed from blacklist successfully", nil)
}

// RecordBehavior 处理记录用户行为的HTTP请求。
func (h *Handler) RecordBehavior(c *gin.Context) {
	var req struct {
		UserID   uint64 `json:"user_id" binding:"required"`
		IP       string `json:"ip" binding:"required"`
		DeviceID string `json:"device_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ctx := ctxutil.WithUserAgent(c.Request.Context(), c.Request.UserAgent())
	if err := h.app.RecordUserBehavior(ctx, req.UserID, req.IP, req.DeviceID); err != nil {
		h.logger.ErrorContext(ctx, "failed to record behavior", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to record behavior", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Behavior recorded successfully", nil)
}

// GetRiskAnalysisResult 处理获取风险分析结果的请求。
func (h *Handler) GetRiskAnalysisResult(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	result, err := h.app.GetRiskAnalysisResult(c.Request.Context(), userID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get risk analysis result", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get risk analysis result", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Risk analysis result retrieved", result)
}

// CheckBlacklist 处理检查值是否在黑名单中的请求。
func (h *Handler) CheckBlacklist(c *gin.Context) {
	bType := c.Query("type")
	value := c.Query("value")

	if bType == "" || value == "" {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Missing type or value", "type and value query parameters are required")
		return
	}

	blacklist, err := h.app.CheckBlacklist(c.Request.Context(), bType, value)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to check blacklist", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to check blacklist", err.Error())
		return
	}

	if blacklist == nil {
		response.SuccessWithStatus(c, http.StatusOK, "Not blacklisted", gin.H{"blacklisted": false})
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Blacklisted", gin.H{"blacklisted": true, "entry": blacklist})
}

// GetUserBehavior 处理获取用户行为的请求。
func (h *Handler) GetUserBehavior(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	behavior, err := h.app.GetUserBehavior(c.Request.Context(), userID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get user behavior", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get user behavior", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "User behavior retrieved", behavior)
}

// RegisterRoutes 在给定的Gin路由组中注册RiskSecurity模块的HTTP路由。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/risk")
	{
		group.POST("/evaluate", h.EvaluateRisk)
		group.POST("/blacklist", h.AddToBlacklist)
		group.GET("/blacklist", h.CheckBlacklist)
		group.DELETE("/blacklist/:id", h.RemoveFromBlacklist)
		group.POST("/behavior", h.RecordBehavior)
		group.GET("/behavior/:user_id", h.GetUserBehavior)
		group.GET("/result/:user_id", h.GetRiskAnalysisResult)
	}
}

package http

import (
	"net/http"
	"strconv"
	"time"

	"ecommerce/internal/risk_security/application"
	"ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.RiskService
	logger  *slog.Logger
}

func NewHandler(service *application.RiskService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// EvaluateRisk 评估风险
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

	result, err := h.service.EvaluateRisk(c.Request.Context(), req.UserID, req.IP, req.DeviceID, req.Amount)
	if err != nil {
		h.logger.Error("Failed to evaluate risk", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to evaluate risk", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Risk evaluated successfully", result)
}

// AddToBlacklist 添加黑名单
func (h *Handler) AddToBlacklist(c *gin.Context) {
	var req struct {
		Type     string `json:"type" binding:"required"`
		Value    string `json:"value" binding:"required"`
		Reason   string `json:"reason"`
		Duration string `json:"duration"` // e.g., "24h"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		duration = 24 * time.Hour // Default
	}

	if err := h.service.AddToBlacklist(c.Request.Context(), req.Type, req.Value, req.Reason, duration); err != nil {
		h.logger.Error("Failed to add to blacklist", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add to blacklist", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Added to blacklist successfully", nil)
}

// RemoveFromBlacklist 移除黑名单
func (h *Handler) RemoveFromBlacklist(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.RemoveFromBlacklist(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to remove from blacklist", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to remove from blacklist", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Removed from blacklist successfully", nil)
}

// RecordBehavior 记录行为
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

	if err := h.service.RecordUserBehavior(c.Request.Context(), req.UserID, req.IP, req.DeviceID); err != nil {
		h.logger.Error("Failed to record behavior", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to record behavior", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Behavior recorded successfully", nil)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/risk")
	{
		group.POST("/evaluate", h.EvaluateRisk)
		group.POST("/blacklist", h.AddToBlacklist)
		group.DELETE("/blacklist/:id", h.RemoveFromBlacklist)
		group.POST("/behavior", h.RecordBehavior)
	}
}

package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/user_tier/application"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.UserTierService
	logger  *slog.Logger
}

func NewHandler(service *application.UserTierService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetTier 获取等级
func (h *Handler) GetTier(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	tier, err := h.service.GetUserTier(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user tier", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get user tier", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "User tier retrieved successfully", tier)
}

// GetPoints 获取积分
func (h *Handler) GetPoints(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	points, err := h.service.GetPoints(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get points", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get points", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Points retrieved successfully", gin.H{"points": points})
}

// Exchange 兑换
func (h *Handler) Exchange(c *gin.Context) {
	var req struct {
		UserID     uint64 `json:"user_id" binding:"required"`
		ExchangeID uint64 `json:"exchange_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.Exchange(c.Request.Context(), req.UserID, req.ExchangeID); err != nil {
		h.logger.Error("Failed to exchange", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to exchange", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Exchange successful", nil)
}

// ListExchanges 兑换列表
func (h *Handler) ListExchanges(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListExchanges(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list exchanges", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list exchanges", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Exchanges listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListPointsLogs 积分日志
func (h *Handler) ListPointsLogs(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListPointsLogs(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list points logs", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list points logs", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Points logs listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/user_tier")
	{
		group.GET("/:user_id", h.GetTier)
		group.GET("/:user_id/points", h.GetPoints)
		group.GET("/:user_id/points/logs", h.ListPointsLogs)
		group.POST("/exchange", h.Exchange)
		group.GET("/exchanges", h.ListExchanges)
	}
}

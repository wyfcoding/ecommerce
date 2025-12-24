package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/multichannel/application"
	"github.com/wyfcoding/ecommerce/internal/multichannel/domain"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 处理 HTTP 或 gRPC 请求。
type Handler struct {
	service *application.MultiChannelService
	logger  *slog.Logger
}

// NewHandler 处理 HTTP 或 gRPC 请求。
func NewHandler(service *application.MultiChannelService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) RegisterChannel(c *gin.Context) {
	var req domain.Channel
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.RegisterChannel(c.Request.Context(), &req); err != nil {
		h.logger.Error("Failed to register channel", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to register channel", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Channel registered successfully", req)
}

func (h *Handler) SyncOrders(c *gin.Context) {
	var req struct {
		ChannelID uint64 `json:"channel_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.SyncOrders(c.Request.Context(), req.ChannelID); err != nil {
		h.logger.Error("Failed to sync orders", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to sync orders", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Orders synced successfully", nil)
}

func (h *Handler) ListChannels(c *gin.Context) {
	channels, err := h.service.ListChannels(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list channels", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list channels", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Channels listed successfully", channels)
}

func (h *Handler) ListOrders(c *gin.Context) {
	channelID, _ := strconv.ParseUint(c.Query("channel_id"), 10, 64)
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListOrders(c.Request.Context(), channelID, status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list orders", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list orders", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Orders listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/multi-channel")
	{
		group.POST("/channels", h.RegisterChannel)
		group.GET("/channels", h.ListChannels)
		group.POST("/sync/orders", h.SyncOrders)
		group.GET("/orders", h.ListOrders)
	}
}

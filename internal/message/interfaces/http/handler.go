package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/message/application"
	"github.com/wyfcoding/ecommerce/internal/message/domain"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 处理 HTTP 或 gRPC 请求。
type Handler struct {
	service *application.MessageService
	logger  *slog.Logger
}

// NewHandler 处理 HTTP 或 gRPC 请求。
func NewHandler(service *application.MessageService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) SendMessage(c *gin.Context) {
	var req struct {
		SenderID    uint64 `json:"sender_id"`
		ReceiverID  uint64 `json:"receiver_id" binding:"required"`
		MessageType string `json:"message_type" binding:"required"`
		Title       string `json:"title" binding:"required"`
		Content     string `json:"content" binding:"required"`
		Link        string `json:"link"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	senderID := req.SenderID

	message, err := h.service.SendMessage(c.Request.Context(), senderID, req.ReceiverID, domain.MessageType(req.MessageType), req.Title, req.Content, req.Link)
	if err != nil {
		h.logger.Error("Failed to send message", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to send message", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Message sent successfully", message)
}

func (h *Handler) ListMessages(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	statusStr := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var status *int
	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil {
			status = &s
		}
	}

	list, total, err := h.service.ListMessages(c.Request.Context(), userID, status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list messages", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list messages", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Messages listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) MarkAsRead(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.MarkAsRead(c.Request.Context(), id, req.UserID); err != nil {
		h.logger.Error("Failed to mark message as read", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to mark message as read", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Message marked as read successfully", nil)
}

func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	count, err := h.service.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get unread count", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get unread count", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Unread count retrieved successfully", gin.H{"count": count})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/messages")
	{
		group.POST("", h.SendMessage)
		group.GET("", h.ListMessages)
		group.PUT("/:id/read", h.MarkAsRead)
		group.GET("/unread-count", h.GetUnreadCount)
	}
}

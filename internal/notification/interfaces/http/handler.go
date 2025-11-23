package http

import (
	"net/http"
	"strconv"

	"ecommerce/internal/notification/application"
	"ecommerce/internal/notification/domain/entity"
	"ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.NotificationService
	logger  *slog.Logger
}

func NewHandler(service *application.NotificationService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// SendNotification 发送通知
func (h *Handler) SendNotification(c *gin.Context) {
	var req struct {
		UserID    uint64                 `json:"user_id" binding:"required"`
		NotifType string                 `json:"notif_type" binding:"required"`
		Channel   string                 `json:"channel" binding:"required"`
		Title     string                 `json:"title" binding:"required"`
		Content   string                 `json:"content" binding:"required"`
		Data      map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	notification, err := h.service.SendNotification(c.Request.Context(), req.UserID, entity.NotificationType(req.NotifType), entity.NotificationChannel(req.Channel), req.Title, req.Content, req.Data)
	if err != nil {
		h.logger.Error("Failed to send notification", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to send notification", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Notification sent successfully", notification)
}

// SendNotificationByTemplate 使用模板发送通知
func (h *Handler) SendNotificationByTemplate(c *gin.Context) {
	var req struct {
		UserID       uint64                 `json:"user_id" binding:"required"`
		TemplateCode string                 `json:"template_code" binding:"required"`
		Variables    map[string]string      `json:"variables"`
		Data         map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	notification, err := h.service.SendNotificationByTemplate(c.Request.Context(), req.UserID, req.TemplateCode, req.Variables, req.Data)
	if err != nil {
		h.logger.Error("Failed to send notification by template", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to send notification by template", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Notification sent successfully", notification)
}

// ListNotifications 获取通知列表
func (h *Handler) ListNotifications(c *gin.Context) {
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

	list, total, err := h.service.ListNotifications(c.Request.Context(), userID, status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list notifications", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list notifications", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Notifications listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// MarkAsRead 标记已读
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
		h.logger.Error("Failed to mark notification as read", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to mark notification as read", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Notification marked as read successfully", nil)
}

// GetUnreadCount 获取未读数
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

// CreateTemplate 创建模板
func (h *Handler) CreateTemplate(c *gin.Context) {
	var req entity.NotificationTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.CreateTemplate(c.Request.Context(), &req); err != nil {
		h.logger.Error("Failed to create template", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create template", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Template created successfully", req)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/notifications")
	{
		group.POST("", h.SendNotification)
		group.POST("/template", h.SendNotificationByTemplate)
		group.GET("", h.ListNotifications)
		group.PUT("/:id/read", h.MarkAsRead)
		group.GET("/unread-count", h.GetUnreadCount)
		group.POST("/templates", h.CreateTemplate)
	}
}

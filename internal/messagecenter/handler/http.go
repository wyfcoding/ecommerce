package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/messagecenter/service"
	"ecommerce/pkg/response"
)

// MessageCenterHandler 消息中心HTTP处理器
type MessageCenterHandler struct {
	service service.MessageCenterService
	logger  *zap.Logger
}

// NewMessageCenterHandler 创建消息中心HTTP处理器
func NewMessageCenterHandler(service service.MessageCenterService, logger *zap.Logger) *MessageCenterHandler {
	return &MessageCenterHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册路由
func (h *MessageCenterHandler) RegisterRoutes(r *gin.RouterGroup) {
	messages := r.Group("/messages")
	{
		// 用户消息
		messages.GET("", h.GetUserMessages)
		messages.GET("/:id", h.GetUserMessage)
		messages.POST("/read", h.MarkAsRead)
		messages.POST("/read-all", h.MarkAllAsRead)
		messages.DELETE("", h.DeleteMessages)
		messages.GET("/unread-count", h.GetUnreadCount)
		messages.GET("/statistics", h.GetStatistics)
		
		// 消息配置
		messages.GET("/config", h.GetUserConfig)
		messages.PUT("/config", h.UpdateUserConfig)
	}
}

// GetUserMessages 获取用户消息列表
func (h *MessageCenterHandler) GetUserMessages(c *gin.Context) {
	userID := c.GetUint64("userID")
	messageType := c.Query("type")
	status := c.Query("status")
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))

	messages, total, err := h.service.GetUserMessages(c.Request.Context(), userID, messageType, status, int32(pageSize), int32(pageNum))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取消息列表失败", err)
		return
	}

	response.SuccessWithPagination(c, messages, total, int32(pageNum), int32(pageSize))
}

// GetUserMessage 获取消息详情
func (h *MessageCenterHandler) GetUserMessage(c *gin.Context) {
	userID := c.GetUint64("userID")
	messageID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	message, err := h.service.GetUserMessage(c.Request.Context(), userID, messageID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "消息不存在", err)
		return
	}

	response.Success(c, message)
}

// MarkAsRead 标记为已读
func (h *MessageCenterHandler) MarkAsRead(c *gin.Context) {
	userID := c.GetUint64("userID")
	
	var req struct {
		MessageIDs []uint64 `json:"messageIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := h.service.MarkAsRead(c.Request.Context(), userID, req.MessageIDs); err != nil {
		response.Error(c, http.StatusInternalServerError, "标记已读失败", err)
		return
	}

	response.Success(c, nil)
}

// MarkAllAsRead 标记全部已读
func (h *MessageCenterHandler) MarkAllAsRead(c *gin.Context) {
	userID := c.GetUint64("userID")
	messageType := c.Query("type")

	if err := h.service.MarkAllAsRead(c.Request.Context(), userID, messageType); err != nil {
		response.Error(c, http.StatusInternalServerError, "标记全部已读失败", err)
		return
	}

	response.Success(c, nil)
}

// DeleteMessages 删除消息
func (h *MessageCenterHandler) DeleteMessages(c *gin.Context) {
	userID := c.GetUint64("userID")
	
	var req struct {
		MessageIDs []uint64 `json:"messageIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := h.service.DeleteUserMessage(c.Request.Context(), userID, req.MessageIDs); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除消息失败", err)
		return
	}

	response.Success(c, nil)
}

// GetUnreadCount 获取未读消息数
func (h *MessageCenterHandler) GetUnreadCount(c *gin.Context) {
	userID := c.GetUint64("userID")

	count, err := h.service.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取未读数失败", err)
		return
	}

	response.Success(c, gin.H{"count": count})
}

// GetStatistics 获取消息统计
func (h *MessageCenterHandler) GetStatistics(c *gin.Context) {
	userID := c.GetUint64("userID")

	stats, err := h.service.GetStatistics(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取统计失败", err)
		return
	}

	response.Success(c, stats)
}

// GetUserConfig 获取用户消息配置
func (h *MessageCenterHandler) GetUserConfig(c *gin.Context) {
	userID := c.GetUint64("userID")

	config, err := h.service.GetUserConfig(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取配置失败", err)
		return
	}

	response.Success(c, config)
}

// UpdateUserConfig 更新用户消息配置
func (h *MessageCenterHandler) UpdateUserConfig(c *gin.Context) {
	userID := c.GetUint64("userID")
	
	var req struct {
		SystemEnabled      bool   `json:"systemEnabled"`
		OrderEnabled       bool   `json:"orderEnabled"`
		PromotionEnabled   bool   `json:"promotionEnabled"`
		ActivityEnabled    bool   `json:"activityEnabled"`
		NoticeEnabled      bool   `json:"noticeEnabled"`
		InteractionEnabled bool   `json:"interactionEnabled"`
		PushEnabled        bool   `json:"pushEnabled"`
		EmailEnabled       bool   `json:"emailEnabled"`
		SMSEnabled         bool   `json:"smsEnabled"`
		DoNotDisturbStart  string `json:"doNotDisturbStart"`
		DoNotDisturbEnd    string `json:"doNotDisturbEnd"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	config, err := h.service.GetUserConfig(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取配置失败", err)
		return
	}

	// 更新配置
	config.SystemEnabled = req.SystemEnabled
	config.OrderEnabled = req.OrderEnabled
	config.PromotionEnabled = req.PromotionEnabled
	config.ActivityEnabled = req.ActivityEnabled
	config.NoticeEnabled = req.NoticeEnabled
	config.InteractionEnabled = req.InteractionEnabled
	config.PushEnabled = req.PushEnabled
	config.EmailEnabled = req.EmailEnabled
	config.SMSEnabled = req.SMSEnabled
	config.DoNotDisturbStart = req.DoNotDisturbStart
	config.DoNotDisturbEnd = req.DoNotDisturbEnd

	if err := h.service.UpdateUserConfig(c.Request.Context(), config); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新配置失败", err)
		return
	}

	response.Success(c, config)
}

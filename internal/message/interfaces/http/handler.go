package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/message/application"   // 导入消息模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/message/domain/entity" // 导入消息模块的领域实体。
	"github.com/wyfcoding/pkg/response"                             // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了Message模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.MessageService // 依赖Message应用服务，处理核心业务逻辑。
	logger  *slog.Logger                // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Message HTTP Handler 实例。
func NewHandler(service *application.MessageService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// SendMessage 处理发送消息的HTTP请求。
// Method: POST
// Path: /messages
func (h *Handler) SendMessage(c *gin.Context) {
	// 定义请求体结构，用于接收消息发送信息。
	var req struct {
		SenderID    uint64 `json:"sender_id"`                       // 发送者ID，如果不是系统消息，通常从认证上下文获取。
		ReceiverID  uint64 `json:"receiver_id" binding:"required"`  // 接收者ID，必填。
		MessageType string `json:"message_type" binding:"required"` // 消息类型，必填。
		Title       string `json:"title" binding:"required"`        // 标题，必填。
		Content     string `json:"content" binding:"required"`      // 内容，必填。
		Link        string `json:"link"`                            // 链接，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// TODO: 如果发送者不是系统，则应该从认证上下文（如JWT）中获取 SenderID，而不是直接从请求体。
	// 当前实现req.SenderID如果为0，需要特殊处理或假设已在前端处理。
	senderID := req.SenderID

	// 将req.MessageType（字符串）转换为实体MessageType。
	// 注意：这里进行了直接转换，如果req.MessageType是未知类型，可能导致错误或默认值。
	// 实际应用中，可能需要增加验证逻辑或映射函数。
	message, err := h.service.SendMessage(c.Request.Context(), senderID, req.ReceiverID, entity.MessageType(req.MessageType), req.Title, req.Content, req.Link)
	if err != nil {
		h.logger.Error("Failed to send message", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to send message", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Message sent successfully", message)
}

// ListMessages 处理获取用户消息列表的HTTP请求。
// Method: GET
// Path: /messages
func (h *Handler) ListMessages(c *gin.Context) {
	// 从查询参数中获取用户ID。
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	// 从查询参数中获取消息状态字符串，并尝试转换为 int 类型。
	statusStr := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var status *int
	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil { // 只有当状态字符串能成功转换为int时才设置过滤状态。
			status = &s
		}
	}

	// 调用应用服务层获取消息列表。
	list, total, err := h.service.ListMessages(c.Request.Context(), userID, status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list messages", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list messages", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Messages listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// MarkAsRead 处理标记消息为已读的HTTP请求。
// Method: PUT
// Path: /messages/:id/read
func (h *Handler) MarkAsRead(c *gin.Context) {
	// 从URL路径中解析消息ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 定义请求体结构，用于接收用户ID。
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"` // 用户ID，必填。
	}
	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层标记消息为已读。
	if err := h.service.MarkAsRead(c.Request.Context(), id, req.UserID); err != nil {
		h.logger.Error("Failed to mark message as read", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to mark message as read", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Message marked as read successfully", nil)
}

// GetUnreadCount 处理获取用户未读消息数量的HTTP请求。
// Method: GET
// Path: /messages/unread-count
func (h *Handler) GetUnreadCount(c *gin.Context) {
	// 从查询参数中获取用户ID。
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	// 调用应用服务层获取未读消息数量。
	count, err := h.service.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get unread count", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get unread count", err.Error())
		return
	}

	// 返回成功的响应，包含未读消息数量。
	response.SuccessWithStatus(c, http.StatusOK, "Unread count retrieved successfully", gin.H{"count": count})
}

// RegisterRoutes 在给定的Gin路由组中注册Message模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /messages 路由组，用于所有消息相关接口。
	group := r.Group("/messages")
	{
		group.POST("", h.SendMessage)                // 发送消息。
		group.GET("", h.ListMessages)                // 获取消息列表。
		group.PUT("/:id/read", h.MarkAsRead)         // 标记消息为已读。
		group.GET("/unread-count", h.GetUnreadCount) // 获取未读消息数量。
		// TODO: 补充获取消息详情、删除消息、获取会话列表等接口。
	}
}

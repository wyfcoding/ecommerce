package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/notification/application"   // 导入通知模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/notification/domain/entity" // 导入通知模块的领域实体。
	"github.com/wyfcoding/ecommerce/pkg/response"                        // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了Notification模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.NotificationService // 依赖Notification应用服务，处理核心业务逻辑。
	logger  *slog.Logger                     // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Notification HTTP Handler 实例。
func NewHandler(service *application.NotificationService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// SendNotification 处理发送通知的HTTP请求。
// Method: POST
// Path: /notifications
func (h *Handler) SendNotification(c *gin.Context) {
	// 定义请求体结构，用于接收通知发送信息。
	var req struct {
		UserID    uint64                 `json:"user_id" binding:"required"`    // 用户ID，必填。
		NotifType string                 `json:"notif_type" binding:"required"` // 通知类型，必填。
		Channel   string                 `json:"channel" binding:"required"`    // 通知渠道，必填。
		Title     string                 `json:"title" binding:"required"`      // 标题，必填。
		Content   string                 `json:"content" binding:"required"`    // 内容，必填。
		Data      map[string]interface{} `json:"data"`                          // 附加数据，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层发送通知。
	notification, err := h.service.SendNotification(c.Request.Context(), req.UserID, entity.NotificationType(req.NotifType), entity.NotificationChannel(req.Channel), req.Title, req.Content, req.Data)
	if err != nil {
		h.logger.Error("Failed to send notification", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to send notification", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Notification sent successfully", notification)
}

// SendNotificationByTemplate 处理使用模板发送通知的HTTP请求。
// Method: POST
// Path: /notifications/template
func (h *Handler) SendNotificationByTemplate(c *gin.Context) {
	// 定义请求体结构，用于接收模板发送信息。
	var req struct {
		UserID       uint64                 `json:"user_id" binding:"required"`       // 用户ID，必填。
		TemplateCode string                 `json:"template_code" binding:"required"` // 模板代码，必填。
		Variables    map[string]string      `json:"variables"`                        // 模板变量，选填。
		Data         map[string]interface{} `json:"data"`                             // 附加数据，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层使用模板发送通知。
	notification, err := h.service.SendNotificationByTemplate(c.Request.Context(), req.UserID, req.TemplateCode, req.Variables, req.Data)
	if err != nil {
		h.logger.Error("Failed to send notification by template", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to send notification by template", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Notification sent successfully", notification)
}

// ListNotifications 处理获取用户通知列表的HTTP请求。
// Method: GET
// Path: /notifications
func (h *Handler) ListNotifications(c *gin.Context) {
	// 从查询参数中获取用户ID。
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	// 从查询参数中获取状态字符串，并尝试转换为 int 类型。
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

	// 调用应用服务层获取通知列表。
	list, total, err := h.service.ListNotifications(c.Request.Context(), userID, status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list notifications", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list notifications", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Notifications listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// MarkAsRead 处理标记通知为已读的HTTP请求。
// Method: PUT
// Path: /notifications/:id/read
func (h *Handler) MarkAsRead(c *gin.Context) {
	// 从URL路径中解析通知ID。
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

	// 调用应用服务层标记通知为已读。
	if err := h.service.MarkAsRead(c.Request.Context(), id, req.UserID); err != nil {
		h.logger.Error("Failed to mark notification as read", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to mark notification as read", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Notification marked as read successfully", nil)
}

// GetUnreadCount 处理获取用户未读通知数量的HTTP请求。
// Method: GET
// Path: /notifications/unread-count
func (h *Handler) GetUnreadCount(c *gin.Context) {
	// 从查询参数中获取用户ID。
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	// 调用应用服务层获取未读通知数量。
	count, err := h.service.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get unread count", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get unread count", err.Error())
		return
	}

	// 返回成功的响应，包含未读通知数量。
	response.SuccessWithStatus(c, http.StatusOK, "Unread count retrieved successfully", gin.H{"count": count})
}

// CreateTemplate 处理创建通知模板的HTTP请求。
// Method: POST
// Path: /notifications/templates
func (h *Handler) CreateTemplate(c *gin.Context) {
	// 定义请求体结构，使用 entity.NotificationTemplate 结构体直接绑定。
	var req entity.NotificationTemplate
	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建通知模板。
	if err := h.service.CreateTemplate(c.Request.Context(), &req); err != nil {
		h.logger.Error("Failed to create template", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create template", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Template created successfully", req)
}

// RegisterRoutes 在给定的Gin路由组中注册Notification模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /notifications 路由组，用于所有通知相关接口。
	group := r.Group("/notifications")
	{
		group.POST("", h.SendNotification)                    // 发送通知。
		group.POST("/template", h.SendNotificationByTemplate) // 使用模板发送通知。
		group.GET("", h.ListNotifications)                    // 获取通知列表。
		group.PUT("/:id/read", h.MarkAsRead)                  // 标记通知为已读。
		group.GET("/unread-count", h.GetUnreadCount)          // 获取未读通知数量。
		group.POST("/templates", h.CreateTemplate)            // 创建通知模板。
		// TODO: 补充获取通知详情、删除通知、更新模板、获取模板列表等接口。
	}
}

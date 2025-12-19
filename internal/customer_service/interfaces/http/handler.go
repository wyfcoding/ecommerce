package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/customer_service/application"   // 导入客户服务模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/customer_service/domain/entity" // 导入客户服务模块的领域实体。
	"github.com/wyfcoding/pkg/response"                                      // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了CustomerService模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.CustomerService // 依赖CustomerService应用服务，处理核心业务逻辑。
	logger  *slog.Logger                 // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 CustomerService HTTP Handler 实例。
func NewHandler(service *application.CustomerService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateTicket 处理创建工单的HTTP请求。
// Method: POST
// Path: /tickets
func (h *Handler) CreateTicket(c *gin.Context) {
	// 定义请求体结构，用于接收工单的创建信息。
	var req struct {
		UserID      uint64                `json:"user_id" binding:"required"`     // 用户ID，必填。
		Subject     string                `json:"subject" binding:"required"`     // 主题，必填。
		Description string                `json:"description" binding:"required"` // 描述，必填。
		Category    string                `json:"category"`                       // 分类，选填。
		Priority    entity.TicketPriority `json:"priority"`                       // 优先级，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建工单。
	ticket, err := h.service.CreateTicket(c.Request.Context(), req.UserID, req.Subject, req.Description, req.Category, req.Priority)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create ticket", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create ticket", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Ticket created successfully", ticket)
}

// ReplyTicket 处理回复工单的HTTP请求。
// Method: POST
// Path: /tickets/:id/reply
func (h *Handler) ReplyTicket(c *gin.Context) {
	// 从URL路径中解析工单ID。
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 定义请求体结构，用于接收消息内容。
	var req struct {
		SenderID   uint64             `json:"sender_id" binding:"required"`   // 发送者ID，必填。
		SenderType string             `json:"sender_type" binding:"required"` // 发送者类型（user/admin），必填。
		Content    string             `json:"content" binding:"required"`     // 消息内容，必填。
		Type       entity.MessageType `json:"type"`                           // 消息类型，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层回复工单。
	message, err := h.service.ReplyTicket(c.Request.Context(), ticketID, req.SenderID, req.SenderType, req.Content, req.Type)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to reply ticket", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to reply ticket", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Ticket replied successfully", message)
}

// ListTickets 处理列出工单的HTTP请求。
// Method: GET
// Path: /tickets
func (h *Handler) ListTickets(c *gin.Context) {
	// 从查询参数中获取用户ID和状态，并设置默认值。
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	status, _ := strconv.Atoi(c.Query("status"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取工单列表。
	list, total, err := h.service.ListTickets(c.Request.Context(), userID, entity.TicketStatus(status), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list tickets", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list tickets", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Tickets listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListMessages 处理获取工单消息列表的HTTP请求。
// Method: GET
// Path: /tickets/:id/messages
func (h *Handler) ListMessages(c *gin.Context) {
	// 从URL路径中解析工单ID。
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 调用应用服务层获取工单消息列表。
	list, total, err := h.service.ListMessages(c.Request.Context(), ticketID, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list messages", "error", err)
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

// CloseTicket 处理关闭工单的HTTP请求。
// Method: PUT
// Path: /tickets/:id/close
func (h *Handler) CloseTicket(c *gin.Context) {
	// 从URL路径中解析工单ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层关闭工单。
	err = h.service.CloseTicket(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to close ticket", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to close ticket", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Ticket closed successfully", nil)
}

// RegisterRoutes 在给定的Gin路由组中注册CustomerService模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /tickets 路由组，用于所有工单相关接口。
	group := r.Group("/tickets")
	{
		group.POST("", h.CreateTicket)             // 创建工单。
		group.GET("", h.ListTickets)               // 获取工单列表。
		group.GET("/:id/messages", h.ListMessages) // 获取工单消息列表。
		group.POST("/:id/reply", h.ReplyTicket)    // 回复工单。
		group.PUT("/:id/close", h.CloseTicket)     // 关闭工单。
		// TODO: 补充获取工单详情、更新工单、解决工单等接口。
	}
}

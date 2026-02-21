package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/support/application"
	"github.com/wyfcoding/ecommerce/internal/support/domain"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 结构体定义了Customer模块的HTTP处理层。
type Handler struct {
	cmd    *application.SupportCommandService
	query  *application.SupportQueryService
	logger *slog.Logger
}

// NewHandler 创建并返回一个新的 Customer HTTP Handler 实例。
func NewHandler(cmd *application.SupportCommandService, query *application.SupportQueryService, logger *slog.Logger) *Handler {
	return &Handler{
		cmd:    cmd,
		query:  query,
		logger: logger,
	}
}

// CreateTicket 处理创建工单的HTTP请求。
func (h *Handler) CreateTicket(c *gin.Context) {
	var req struct {
		UserID      uint64                `json:"user_id" binding:"required"`
		Subject     string                `json:"subject" binding:"required"`
		Description string                `json:"description" binding:"required"`
		Category    string                `json:"category"`
		Priority    domain.TicketPriority `json:"priority"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ticket, err := h.cmd.CreateTicket(c.Request.Context(), req.UserID, req.Subject, req.Description, req.Category, req.Priority)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to create ticket", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create ticket", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Ticket created successfully", ticket)
}

// ReplyTicket 处理回复工单的HTTP请求。
func (h *Handler) ReplyTicket(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		SenderID   uint64             `json:"sender_id" binding:"required"`
		SenderType string             `json:"sender_type" binding:"required"`
		Content    string             `json:"content" binding:"required"`
		Type       domain.MessageType `json:"type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	message, err := h.cmd.ReplyTicket(c.Request.Context(), ticketID, req.SenderID, req.SenderType, req.Content, req.Type)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to reply ticket", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to reply ticket", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Ticket replied successfully", message)
}

// ListTickets 处理列出工单的HTTP请求。
func (h *Handler) ListTickets(c *gin.Context) {
	var (
		userID   uint64
		status   int
		page     int
		pageSize int
		err      error
	)

	if val := c.Query("user_id"); val != "" {
		userID, err = strconv.ParseUint(val, 10, 64)
		if err != nil {
			response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid user_id", err.Error())
			return
		}
	}

	if val := c.Query("status"); val != "" {
		status, err = strconv.Atoi(val)
		if err != nil {
			response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid status", err.Error())
			return
		}
	}

	page, err = strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err = strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	list, total, err := h.query.ListTickets(c.Request.Context(), userID, domain.TicketStatus(status), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list tickets", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list tickets", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Tickets listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListMessages 处理获取工单消息列表的HTTP请求。
func (h *Handler) ListMessages(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil || pageSize <= 0 {
		pageSize = 20
	}

	list, total, err := h.query.ListMessages(c.Request.Context(), ticketID, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list messages", "error", err)
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

// CloseTicket 处理关闭工单的HTTP请求。
func (h *Handler) CloseTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	err = h.cmd.CloseTicket(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to close ticket", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to close ticket", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Ticket closed successfully", nil)
}

// SearchKnowledgeArticles 处理知识库搜索请求。
func (h *Handler) SearchKnowledgeArticles(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Query parameter 'q' is required", "")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	articles, err := h.query.SearchKnowledgeArticles(c.Request.Context(), query, limit)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to search knowledge articles", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to search knowledge articles", err.Error())
		return
	}

	response.SuccessWithMsg(c, "Knowledge articles retrieved successfully", articles)
}

// GetChatbotResponse 处理 AI 客服会话请求。
func (h *Handler) GetChatbotResponse(c *gin.Context) {
	var req struct {
		ConversationID string `json:"conversation_id" binding:"required"`
		Message        string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	result, err := h.cmd.GetChatbotResponse(c.Request.Context(), req.ConversationID, req.Message)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get chatbot response", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get chatbot response", err.Error())
		return
	}

	response.SuccessWithMsg(c, "AI response generated successfully", result)
}

// RegisterRoutes 在给定的Gin路由组中注册Customer模块的HTTP路由。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/tickets")
	{
		group.POST("", h.CreateTicket)
		group.GET("", h.ListTickets)
		group.GET("/:id/messages", h.ListMessages)
		group.POST("/:id/reply", h.ReplyTicket)
		group.PUT("/:id/close", h.CloseTicket)
	}

	ai := r.Group("/ai")
	{
		ai.GET("/search", h.SearchKnowledgeArticles)
		ai.POST("/chat", h.GetChatbotResponse)
	}
}

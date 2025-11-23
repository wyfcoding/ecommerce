package http

import (
	"net/http"
	"strconv"

	"ecommerce/internal/customer_service/application"
	"ecommerce/internal/customer_service/domain/entity"
	"ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.CustomerService
	logger  *slog.Logger
}

func NewHandler(service *application.CustomerService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateTicket 创建工单
func (h *Handler) CreateTicket(c *gin.Context) {
	var req struct {
		UserID      uint64                `json:"user_id" binding:"required"`
		Subject     string                `json:"subject" binding:"required"`
		Description string                `json:"description" binding:"required"`
		Category    string                `json:"category"`
		Priority    entity.TicketPriority `json:"priority"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ticket, err := h.service.CreateTicket(c.Request.Context(), req.UserID, req.Subject, req.Description, req.Category, req.Priority)
	if err != nil {
		h.logger.Error("Failed to create ticket", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create ticket", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Ticket created successfully", ticket)
}

// ReplyTicket 回复工单
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
		Type       entity.MessageType `json:"type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	message, err := h.service.ReplyTicket(c.Request.Context(), ticketID, req.SenderID, req.SenderType, req.Content, req.Type)
	if err != nil {
		h.logger.Error("Failed to reply ticket", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to reply ticket", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Ticket replied successfully", message)
}

// ListTickets 获取工单列表
func (h *Handler) ListTickets(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	status, _ := strconv.Atoi(c.Query("status"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListTickets(c.Request.Context(), userID, entity.TicketStatus(status), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list tickets", "error", err)
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

// ListMessages 获取工单消息
func (h *Handler) ListMessages(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.service.ListMessages(c.Request.Context(), ticketID, page, pageSize)
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

// CloseTicket 关闭工单
func (h *Handler) CloseTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	err = h.service.CloseTicket(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to close ticket", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to close ticket", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Ticket closed successfully", nil)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/tickets")
	{
		group.POST("", h.CreateTicket)
		group.GET("", h.ListTickets)
		group.GET("/:id/messages", h.ListMessages)
		group.POST("/:id/reply", h.ReplyTicket)
		group.PUT("/:id/close", h.CloseTicket)
	}
}

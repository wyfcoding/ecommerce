package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/aftersales/application"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *application.AfterSalesService
	logger  *slog.Logger
}

func NewHandler(service *application.AfterSalesService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) Create(c *gin.Context) {
	var req struct {
		OrderID     uint64                   `json:"order_id" binding:"required"`
		OrderNo     string                   `json:"order_no" binding:"required"`
		UserID      uint64                   `json:"user_id" binding:"required"`
		Type        domain.AfterSalesType    `json:"type" binding:"required"`
		Reason      string                   `json:"reason" binding:"required"`
		Description string                   `json:"description"`
		Images      []string                 `json:"images"`
		Items       []*domain.AfterSalesItem `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid data: "+err.Error())
		return
	}

	afterSales, err := h.service.CreateAfterSales(c.Request.Context(), req.OrderID, req.OrderNo, req.UserID, req.Type, req.Reason, req.Description, req.Images, req.Items)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create after-sales", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Created", afterSales)
}

func (h *Handler) Approve(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid aftersales ID")
		return
	}

	var req struct {
		Operator string `json:"operator" binding:"required"`
		Amount   int64  `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid input")
		return
	}

	if err := h.service.Approve(c.Request.Context(), id, req.Operator, req.Amount); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to approve after-sales", "id", id, "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) Reject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid aftersales ID")
		return
	}

	var req struct {
		Operator string `json:"operator" binding:"required"`
		Reason   string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid input")
		return
	}

	if err := h.service.Reject(c.Request.Context(), id, req.Operator, req.Reason); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to reject after-sales", "id", id, "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) List(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		response.BadRequest(c, "invalid page")
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		response.BadRequest(c, "invalid page_size")
		return
	}
	userID, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)

	query := &domain.AfterSalesQuery{
		Page:     page,
		PageSize: pageSize,
		UserID:   userID,
	}

	list, total, err := h.service.List(c.Request.Context(), query)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list after-sales", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (h *Handler) GetDetails(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid ID")
		return
	}

	details, err := h.service.GetDetails(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get after-sales details", "id", id, "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, details)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/aftersales")
	{
		group.POST("", h.Create)
		group.GET("", h.List)
		group.GET("/:id", h.GetDetails)
		group.POST("/:id/approve", h.Approve)
		group.POST("/:id/reject", h.Reject)
	}
}

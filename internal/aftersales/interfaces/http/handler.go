package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/aftersales/application"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain/repository"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
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

// Create 创建售后
func (h *Handler) Create(c *gin.Context) {
	var req struct {
		OrderID     uint64                   `json:"order_id" binding:"required"`
		OrderNo     string                   `json:"order_no" binding:"required"`
		UserID      uint64                   `json:"user_id" binding:"required"`
		Type        entity.AfterSalesType    `json:"type" binding:"required"`
		Reason      string                   `json:"reason" binding:"required"`
		Description string                   `json:"description"`
		Images      []string                 `json:"images"`
		Items       []*entity.AfterSalesItem `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	afterSales, err := h.service.CreateAfterSales(c.Request.Context(), req.OrderID, req.OrderNo, req.UserID, req.Type, req.Reason, req.Description, req.Images, req.Items)
	if err != nil {
		h.logger.Error("Failed to create after-sales", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create after-sales", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "After-sales created successfully", afterSales)
}

// Approve 批准
func (h *Handler) Approve(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		Operator string `json:"operator" binding:"required"`
		Amount   int64  `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.Approve(c.Request.Context(), id, req.Operator, req.Amount); err != nil {
		h.logger.Error("Failed to approve after-sales", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to approve after-sales", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "After-sales approved successfully", nil)
}

// Reject 拒绝
func (h *Handler) Reject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		Operator string `json:"operator" binding:"required"`
		Reason   string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.Reject(c.Request.Context(), id, req.Operator, req.Reason); err != nil {
		h.logger.Error("Failed to reject after-sales", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to reject after-sales", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "After-sales rejected successfully", nil)
}

// List 列表
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	userID, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)

	query := &repository.AfterSalesQuery{
		Page:     page,
		PageSize: pageSize,
		UserID:   userID,
	}

	list, total, err := h.service.List(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to list after-sales", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list after-sales", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "After-sales listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetDetails 详情
func (h *Handler) GetDetails(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	details, err := h.service.GetDetails(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get after-sales details", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get after-sales details", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "After-sales details retrieved successfully", details)
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

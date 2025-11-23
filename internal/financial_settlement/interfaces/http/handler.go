package http

import (
	"net/http"
	"strconv"
	"time"

	"ecommerce/internal/financial_settlement/application"
	"ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.FinancialSettlementService
	logger  *slog.Logger
}

func NewHandler(service *application.FinancialSettlementService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateSettlement 创建结算单
func (h *Handler) CreateSettlement(c *gin.Context) {
	var req struct {
		SellerID  uint64    `json:"seller_id" binding:"required"`
		Period    string    `json:"period" binding:"required"`
		StartDate time.Time `json:"start_date" binding:"required"`
		EndDate   time.Time `json:"end_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	settlement, err := h.service.CreateSettlement(c.Request.Context(), req.SellerID, req.Period, req.StartDate, req.EndDate)
	if err != nil {
		h.logger.Error("Failed to create settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create settlement", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Settlement created successfully", settlement)
}

// ApproveSettlement 审核结算单
func (h *Handler) ApproveSettlement(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		ApprovedBy string `json:"approved_by" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.ApproveSettlement(c.Request.Context(), id, req.ApprovedBy); err != nil {
		h.logger.Error("Failed to approve settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to approve settlement", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Settlement approved successfully", nil)
}

// RejectSettlement 拒绝结算单
func (h *Handler) RejectSettlement(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.RejectSettlement(c.Request.Context(), id, req.Reason); err != nil {
		h.logger.Error("Failed to reject settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to reject settlement", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Settlement rejected successfully", nil)
}

// GetSettlement 获取结算单详情
func (h *Handler) GetSettlement(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	settlement, err := h.service.GetSettlement(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get settlement", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Settlement retrieved successfully", settlement)
}

// ListSettlements 获取结算单列表
func (h *Handler) ListSettlements(c *gin.Context) {
	sellerID, _ := strconv.ParseUint(c.Query("seller_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListSettlements(c.Request.Context(), sellerID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list settlements", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list settlements", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Settlements listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ProcessPayment 处理支付
func (h *Handler) ProcessPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	payment, err := h.service.ProcessPayment(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to process payment", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to process payment", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Payment processed successfully", payment)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/settlements")
	{
		group.POST("", h.CreateSettlement)
		group.GET("/:id", h.GetSettlement)
		group.GET("", h.ListSettlements)
		group.POST("/:id/approve", h.ApproveSettlement)
		group.POST("/:id/reject", h.RejectSettlement)
		group.POST("/:id/pay", h.ProcessPayment)
	}
}

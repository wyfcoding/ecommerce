package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/financialsettlement/application"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 处理 HTTP 或 gRPC 请求。
type Handler struct {
	app    *application.SettlementService
	logger *slog.Logger
}

// NewHandler 创建并返回一个新的 FinancialSettlement HTTP Handler 实例。
func NewHandler(app *application.SettlementService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// CreateSettlement 处理创建结算单的HTTP请求。
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

	settlement, err := h.app.CreateSettlement(c.Request.Context(), req.SellerID, req.Period, req.StartDate, req.EndDate)
	if err != nil {
		h.logger.Error("Failed to create settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create settlement", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Settlement created successfully", settlement)
}

// ApproveSettlement 处理审核批准结算单的HTTP请求。
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

	if err := h.app.ApproveSettlement(c.Request.Context(), id, req.ApprovedBy); err != nil {
		h.logger.Error("Failed to approve settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to approve settlement", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Settlement approved successfully", nil)
}

// RejectSettlement 处理审核拒绝结算单的HTTP请求。
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

	if err := h.app.RejectSettlement(c.Request.Context(), id, req.Reason); err != nil {
		h.logger.Error("Failed to reject settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to reject settlement", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Settlement rejected successfully", nil)
}

// GetSettlement 处理获取结算单详情的HTTP请求。
func (h *Handler) GetSettlement(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	settlement, err := h.app.GetSettlement(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get settlement", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Settlement retrieved successfully", settlement)
}

// ListSettlements 处理获取结算单列表的HTTP请求。
func (h *Handler) ListSettlements(c *gin.Context) {
	sellerID, _ := strconv.ParseUint(c.Query("seller_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.app.ListSettlements(c.Request.Context(), sellerID, page, pageSize)
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

// ProcessPayment 处理结算单支付的HTTP请求。
func (h *Handler) ProcessPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	payment, err := h.app.ProcessPayment(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to process payment", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to process payment", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Payment processed successfully", payment)
}

// RegisterRoutes 在给定的Gin路由组中注册FinancialSettlement模块的HTTP路由。
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

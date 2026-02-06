package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/settlement/application"
	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 处理 HTTP 请求。
type Handler struct {
	cmd    *application.SettlementCommandService
	query  *application.SettlementQueryService
	logger *slog.Logger
}

// NewHandler 创建一个新的 Handler 实例。
func NewHandler(cmd *application.SettlementCommandService, query *application.SettlementQueryService, logger *slog.Logger) *Handler {
	return &Handler{
		cmd:    cmd,
		query:  query,
		logger: logger,
	}
}

// CreateSettlement 处理创建结算单的请求。
func (h *Handler) CreateSettlement(c *gin.Context) {
	var req struct {
		MerchantID uint64 `json:"merchant_id" binding:"required"`
		Cycle      string `json:"cycle" binding:"required"`
		StartDate  string `json:"start_date" binding:"required"` // 支持字符串格式 YYYY-MM-DD
		EndDate    string `json:"end_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid start_date format", err.Error())
		return
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid end_date format", err.Error())
		return
	}

	settlement, err := h.cmd.CreateSettlement(c.Request.Context(), req.MerchantID, req.Cycle, start, end)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to create settlement", "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "settlement created successfully", settlement)
}

// AddOrderToSettlement 处理添加订单到结算单的请求。
func (h *Handler) AddOrderToSettlement(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id format", "")
		return
	}

	var req struct {
		OrderID uint64 `json:"order_id" binding:"required"`
		OrderNo string `json:"order_no" binding:"required"`
		Amount  uint64 `json:"amount" binding:"required"`
	}

	if err := h.cmd.AddOrderToSettlement(c.Request.Context(), id, req.OrderID, req.OrderNo, req.Amount); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to add order to settlement", "id", id, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// ProcessSettlement 处理结算单的请求。
func (h *Handler) ProcessSettlement(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id format", "")
		return
	}

	if err := h.cmd.ProcessSettlement(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to process settlement", "id", id, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// CompleteSettlement 完成结算单的请求。
func (h *Handler) CompleteSettlement(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id format", "")
		return
	}

	if err := h.cmd.CompleteSettlement(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to complete settlement", "id", id, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// ListSettlements 分页查询结算单。
func (h *Handler) ListSettlements(c *gin.Context) {
	merchantID, _ := strconv.ParseUint(c.Query("merchant_id"), 10, 64)
	statusStr := c.Query("status")
	var statusPtr *domain.SettlementStatus
	if statusStr != "" {
		st, _ := strconv.Atoi(statusStr)
		s := domain.SettlementStatus(st)
		statusPtr = &s
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.query.ListSettlements(c.Request.Context(), merchantID, statusPtr, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list settlements", "merchant_id", merchantID, "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, list, total, int32(page), int32(pageSize))
}

// GetSettlement 获取结算单详情。
func (h *Handler) GetSettlement(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id format", "")
		return
	}

	settlement, err := h.query.GetSettlement(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get settlement", "id", id, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, settlement)
}

// GetMerchantAccount 获取商户账户背景。
func (h *Handler) GetMerchantAccount(c *gin.Context) {
	merchantID, err := strconv.ParseUint(c.Param("merchant_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid merchant_id", "")
		return
	}

	account, err := h.query.GetMerchantAccount(c.Request.Context(), merchantID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get merchant account", "merchant_id", merchantID, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, account)
}

// RegisterRoutes 注册路由。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/settlement")
	{
		group.POST("", h.CreateSettlement)
		group.POST("/:id/orders", h.AddOrderToSettlement)
		group.POST("/:id/process", h.ProcessSettlement)
		group.POST("/:id/complete", h.CompleteSettlement)
		group.GET("", h.ListSettlements)
		group.GET("/:id", h.GetSettlement)
		group.GET("/accounts/:merchant_id", h.GetMerchantAccount)
	}
}

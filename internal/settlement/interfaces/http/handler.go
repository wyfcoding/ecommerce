package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/settlement/application"
	"github.com/wyfcoding/pkg/response"

	"log/slog"

	"github.com/gin-gonic/gin"
)

// Handler 结构体定义了结算模块的HTTP处理层。
type Handler struct {
	service *application.SettlementService
	logger  *slog.Logger
}

// NewHandler 创建并返回一个新的结算 HTTP Handler 实例。
func NewHandler(service *application.SettlementService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateSettlement 处理创建结算单的HTTP请求。
func (h *Handler) CreateSettlement(c *gin.Context) {
	var req struct {
		MerchantID uint64 `json:"merchant_id" binding:"required"`
		Cycle      string `json:"cycle" binding:"required"`
		StartDate  string `json:"start_date" binding:"required"`
		EndDate    string `json:"end_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid StartDate format", err.Error())
		return
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid EndDate format", err.Error())
		return
	}

	settlement, err := h.service.CreateSettlement(c.Request.Context(), req.MerchantID, req.Cycle, start, end)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create settlement", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Settlement created successfully", settlement)
}

// AddOrder 处理添加订单到结算单的HTTP请求。
func (h *Handler) AddOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		OrderID uint64 `json:"order_id" binding:"required"`
		OrderNo string `json:"order_no" binding:"required"`
		Amount  uint64 `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.AddOrderToSettlement(c.Request.Context(), id, req.OrderID, req.OrderNo, req.Amount); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to add order", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add order", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Order added successfully", nil)
}

// Process 处理结算单的HTTP请求。
func (h *Handler) Process(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.ProcessSettlement(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to process settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to process settlement", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Settlement processing started", nil)
}

// Complete 处理完成结算单的HTTP请求。
func (h *Handler) Complete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.CompleteSettlement(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to complete settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to complete settlement", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Settlement completed successfully", nil)
}

// List 处理获取结算单列表的HTTP请求。
func (h *Handler) List(c *gin.Context) {
	merchantID, _ := strconv.ParseUint(c.Query("merchant_id"), 10, 64)
	statusStr := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var status *int
	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil {
			status = &s
		}
	}

	list, total, err := h.service.ListSettlements(c.Request.Context(), merchantID, status, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list settlements", "error", err)
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

// GetAccount 处理获取商户账户信息的HTTP请求。
func (h *Handler) GetAccount(c *gin.Context) {
	merchantID, err := strconv.ParseUint(c.Param("merchant_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Merchant ID", err.Error())
		return
	}

	account, err := h.service.GetMerchantAccount(c.Request.Context(), merchantID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get merchant account", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get merchant account", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Merchant account retrieved successfully", account)
}

// RegisterRoutes 在给定的Gin路由组中注册结算模块的HTTP路由。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/settlement")
	{
		group.POST("", h.CreateSettlement)
		group.POST("/:id/orders", h.AddOrder)
		group.POST("/:id/process", h.Process)
		group.POST("/:id/complete", h.Complete)
		group.GET("", h.List)
		group.GET("/accounts/:merchant_id", h.GetAccount)
	}
}

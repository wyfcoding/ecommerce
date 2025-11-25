package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/settlement/application"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.SettlementService
	logger  *slog.Logger
}

func NewHandler(service *application.SettlementService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateSettlement 创建结算单
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

	start, _ := time.Parse("2006-01-02", req.StartDate)
	end, _ := time.Parse("2006-01-02", req.EndDate)

	settlement, err := h.service.CreateSettlement(c.Request.Context(), req.MerchantID, req.Cycle, start, end)
	if err != nil {
		h.logger.Error("Failed to create settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create settlement", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Settlement created successfully", settlement)
}

// AddOrder 添加订单
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
		h.logger.Error("Failed to add order", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add order", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Order added successfully", nil)
}

// Process 处理
func (h *Handler) Process(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.ProcessSettlement(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to process settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to process settlement", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Settlement processing started", nil)
}

// Complete 完成
func (h *Handler) Complete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.CompleteSettlement(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to complete settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to complete settlement", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Settlement completed successfully", nil)
}

// List 列表
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

// GetAccount 获取账户
func (h *Handler) GetAccount(c *gin.Context) {
	merchantID, err := strconv.ParseUint(c.Param("merchant_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Merchant ID", err.Error())
		return
	}

	account, err := h.service.GetMerchantAccount(c.Request.Context(), merchantID)
	if err != nil {
		h.logger.Error("Failed to get merchant account", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get merchant account", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Merchant account retrieved successfully", account)
}

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

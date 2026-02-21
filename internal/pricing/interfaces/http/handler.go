package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/pricing/application"
	"github.com/wyfcoding/ecommerce/internal/pricing/domain"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 处理 HTTP 或 gRPC 请求。
type Handler struct {
	cmd    *application.PricingCommandService
	query  *application.PricingQueryService
	logger *slog.Logger
}

// NewHandler 处理 HTTP 或 gRPC 请求。
func NewHandler(cmd *application.PricingCommandService, query *application.PricingQueryService, logger *slog.Logger) *Handler {
	return &Handler{
		cmd:    cmd,
		query:  query,
		logger: logger,
	}
}

func (h *Handler) CreateRule(c *gin.Context) {
	var req domain.PricingRule
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.cmd.CreateRule(c.Request.Context(), &req); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create rule", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create rule", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Rule created successfully", req)
}

func (h *Handler) CalculatePrice(c *gin.Context) {
	var req struct {
		ProductID          uint64 `json:"product_id"`
		SkuID              uint64 `json:"sku_id" binding:"required"`
		BasePrice          int64  `json:"base_price"`
		CurrentStock       int32  `json:"current_stock"`
		TotalStock         int32  `json:"total_stock"`
		DailyDemand        int32  `json:"daily_demand"`
		AverageDailyDemand int32  `json:"average_daily_demand"`
		CompetitorPrice    int64  `json:"competitor_price"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	pricingReq := &domain.PricingRequest{
		SKUID:              req.SkuID,
		BasePrice:          req.BasePrice,
		CurrentStock:       req.CurrentStock,
		TotalStock:         req.TotalStock,
		DailyDemand:        req.DailyDemand,
		AverageDailyDemand: req.AverageDailyDemand,
		CompetitorPrice:    req.CompetitorPrice,
	}

	price, err := h.cmd.CalculateDynamicPrice(c.Request.Context(), pricingReq)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to calculate dynamic price", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to calculate price", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Price calculated successfully", price)
}

func (h *Handler) ListRules(c *gin.Context) {
	var (
		productID uint64
		err       error
	)
	if val := c.Query("product_id"); val != "" {
		productID, err = strconv.ParseUint(val, 10, 64)
		if err != nil {
			response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid product_id", err.Error())
			return
		}
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	list, total, err := h.query.ListRules(c.Request.Context(), productID, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list rules", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list rules", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Rules listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) ListHistory(c *gin.Context) {
	var (
		productID uint64
		skuID     uint64
		err       error
	)
	if val := c.Query("product_id"); val != "" {
		productID, err = strconv.ParseUint(val, 10, 64)
		if err != nil {
			response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid product_id", err.Error())
			return
		}
	}
	if val := c.Query("sku_id"); val != "" {
		skuID, err = strconv.ParseUint(val, 10, 64)
		if err != nil {
			response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid sku_id", err.Error())
			return
		}
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	list, total, err := h.query.ListHistory(c.Request.Context(), productID, skuID, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list history", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list history", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "History listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/pricing")
	{
		group.POST("/rules", h.CreateRule)
		group.GET("/rules", h.ListRules)
		group.POST("/calculate", h.CalculatePrice)
		group.GET("/history", h.ListHistory)
	}
}

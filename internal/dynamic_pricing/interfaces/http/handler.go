package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/application"
	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/domain"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 结构体定义了DynamicPricing模块的HTTP处理层。
type Handler struct {
	app    *application.DynamicPricingService
	logger *slog.Logger
}

// NewHandler 创建并返回一个新的 DynamicPricing HTTP Handler 实例。
func NewHandler(app *application.DynamicPricingService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// CalculatePrice 处理计算动态价格的HTTP请求。
func (h *Handler) CalculatePrice(c *gin.Context) {
	var req domain.PricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	price, err := h.app.CalculatePrice(c.Request.Context(), &req)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to calculate price", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to calculate price", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Price calculated successfully", price)
}

// GetLatestPrice 处理获取指定SKU最新动态价格的HTTP请求。
func (h *Handler) GetLatestPrice(c *gin.Context) {
	skuID, err := strconv.ParseUint(c.Param("sku_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid SKU ID", err.Error())
		return
	}

	price, err := h.app.GetLatestPrice(c.Request.Context(), skuID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get latest price", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get latest price", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Latest price retrieved successfully", price)
}

// SaveStrategy 处理保存（创建或更新）定价策略的HTTP请求。
func (h *Handler) SaveStrategy(c *gin.Context) {
	var strategy domain.PricingStrategy
	if err := c.ShouldBindJSON(&strategy); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.app.SaveStrategy(c.Request.Context(), &strategy); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to save strategy", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to save strategy", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Strategy saved successfully", strategy)
}

// ListStrategies 处理获取定价策略列表的HTTP请求。
func (h *Handler) ListStrategies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.app.ListStrategies(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list strategies", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list strategies", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Strategies listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册DynamicPricing模块的HTTP路由。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/pricing")
	{
		group.POST("/calculate", h.CalculatePrice)
		group.GET("/sku/:sku_id/latest", h.GetLatestPrice)
		group.POST("/strategies", h.SaveStrategy)
		group.GET("/strategies", h.ListStrategies)
	}
}

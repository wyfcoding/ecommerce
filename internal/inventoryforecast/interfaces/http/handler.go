package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/inventoryforecast/application"
	"github.com/wyfcoding/ecommerce/internal/inventoryforecast/domain"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 结构体定义了InventoryForecast模块的HTTP处理层。
type Handler struct {
	app    *application.InventoryForecastService
	logger *slog.Logger
}

// NewHandler 创建并返回一个新的 InventoryForecast HTTP Handler 实例。
func NewHandler(app *application.InventoryForecastService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// GenerateForecast 处理生成销售预测的HTTP请求。
func (h *Handler) GenerateForecast(c *gin.Context) {
	var req struct {
		SKUID uint64 `json:"sku_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	forecast, err := h.app.GenerateForecast(c.Request.Context(), req.SKUID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to generate forecast", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to generate forecast", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Forecast generated successfully", forecast)
}

// GetForecast 处理获取销售预测的HTTP请求。
func (h *Handler) GetForecast(c *gin.Context) {
	skuID, err := strconv.ParseUint(c.Param("sku_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid SKU ID", err.Error())
		return
	}

	forecast, err := h.app.GetForecast(c.Request.Context(), skuID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get forecast", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get forecast", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Forecast retrieved successfully", forecast)
}

// ListWarnings 处理获取库存预警列表的HTTP请求。
func (h *Handler) ListWarnings(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	list, total, err := h.app.ListWarnings(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list warnings", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list warnings", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Warnings listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListSlowMovingItems 处理获取滞销品列表的HTTP请求。
func (h *Handler) ListSlowMovingItems(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	list, total, err := h.app.ListSlowMovingItems(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list slow moving items", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list slow moving items", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Slow moving items listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListReplenishmentSuggestions 处理获取补货建议列表的HTTP请求。
func (h *Handler) ListReplenishmentSuggestions(c *gin.Context) {
	priority := c.Query("priority")
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	list, total, err := h.app.ListReplenishmentSuggestions(c.Request.Context(), priority, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list replenishment suggestions", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list replenishment suggestions", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Replenishment suggestions listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListStockoutRisks 处理获取缺货风险列表的HTTP请求。
func (h *Handler) ListStockoutRisks(c *gin.Context) {
	level := c.Query("level")
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	list, total, err := h.app.ListStockoutRisks(c.Request.Context(), domain.StockoutRiskLevel(level), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list stockout risks", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list stockout risks", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Stockout risks listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册InventoryForecast模块的HTTP路由。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/inventory-forecast")
	{
		group.POST("/forecasts", h.GenerateForecast)
		group.GET("/forecasts/:sku_id", h.GetForecast)
		group.GET("/warnings", h.ListWarnings)
		group.GET("/slow-moving-items", h.ListSlowMovingItems)
		group.GET("/replenishment-suggestions", h.ListReplenishmentSuggestions)
		group.GET("/stockout-risks", h.ListStockoutRisks)
	}
}

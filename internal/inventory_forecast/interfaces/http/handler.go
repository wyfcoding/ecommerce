package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/application"
	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/domain/entity"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.InventoryForecastService
	logger  *slog.Logger
}

func NewHandler(service *application.InventoryForecastService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GenerateForecast 生成预测
func (h *Handler) GenerateForecast(c *gin.Context) {
	var req struct {
		SKUID uint64 `json:"sku_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	forecast, err := h.service.GenerateForecast(c.Request.Context(), req.SKUID)
	if err != nil {
		h.logger.Error("Failed to generate forecast", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to generate forecast", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Forecast generated successfully", forecast)
}

// GetForecast 获取预测
func (h *Handler) GetForecast(c *gin.Context) {
	skuID, err := strconv.ParseUint(c.Param("sku_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid SKU ID", err.Error())
		return
	}

	forecast, err := h.service.GetForecast(c.Request.Context(), skuID)
	if err != nil {
		h.logger.Error("Failed to get forecast", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get forecast", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Forecast retrieved successfully", forecast)
}

// ListWarnings 获取预警
func (h *Handler) ListWarnings(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListWarnings(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list warnings", "error", err)
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

// ListSlowMovingItems 获取滞销品
func (h *Handler) ListSlowMovingItems(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListSlowMovingItems(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list slow moving items", "error", err)
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

// ListReplenishmentSuggestions 获取补货建议
func (h *Handler) ListReplenishmentSuggestions(c *gin.Context) {
	priority := c.Query("priority")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListReplenishmentSuggestions(c.Request.Context(), priority, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list replenishment suggestions", "error", err)
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

// ListStockoutRisks 获取缺货风险
func (h *Handler) ListStockoutRisks(c *gin.Context) {
	level := c.Query("level")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListStockoutRisks(c.Request.Context(), entity.StockoutRiskLevel(level), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list stockout risks", "error", err)
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

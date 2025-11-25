package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/application"
	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/domain/entity"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.DynamicPricingService
	logger  *slog.Logger
}

func NewHandler(service *application.DynamicPricingService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CalculatePrice 计算价格
func (h *Handler) CalculatePrice(c *gin.Context) {
	var req entity.PricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	price, err := h.service.CalculatePrice(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to calculate price", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to calculate price", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Price calculated successfully", price)
}

// GetLatestPrice 获取最新价格
func (h *Handler) GetLatestPrice(c *gin.Context) {
	skuID, err := strconv.ParseUint(c.Param("sku_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid SKU ID", err.Error())
		return
	}

	price, err := h.service.GetLatestPrice(c.Request.Context(), skuID)
	if err != nil {
		h.logger.Error("Failed to get latest price", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get latest price", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Latest price retrieved successfully", price)
}

// SaveStrategy 保存策略
func (h *Handler) SaveStrategy(c *gin.Context) {
	var strategy entity.PricingStrategy
	if err := c.ShouldBindJSON(&strategy); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.SaveStrategy(c.Request.Context(), &strategy); err != nil {
		h.logger.Error("Failed to save strategy", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to save strategy", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Strategy saved successfully", strategy)
}

// ListStrategies 获取策略列表
func (h *Handler) ListStrategies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListStrategies(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list strategies", "error", err)
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

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/pricing")
	{
		group.POST("/calculate", h.CalculatePrice)
		group.GET("/sku/:sku_id/latest", h.GetLatestPrice)
		group.POST("/strategies", h.SaveStrategy)
		group.GET("/strategies", h.ListStrategies)
	}
}

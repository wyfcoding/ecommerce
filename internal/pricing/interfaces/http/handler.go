package http

import (
	"net/http"
	"strconv"

	"ecommerce/internal/pricing/application"
	"ecommerce/internal/pricing/domain/entity"
	"ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.PricingService
	logger  *slog.Logger
}

func NewHandler(service *application.PricingService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateRule 创建规则
func (h *Handler) CreateRule(c *gin.Context) {
	var req entity.PricingRule
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.CreateRule(c.Request.Context(), &req); err != nil {
		h.logger.Error("Failed to create rule", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create rule", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Rule created successfully", req)
}

// CalculatePrice 计算价格
func (h *Handler) CalculatePrice(c *gin.Context) {
	var req struct {
		ProductID   uint64  `json:"product_id" binding:"required"`
		SkuID       uint64  `json:"sku_id" binding:"required"`
		Demand      float64 `json:"demand"`
		Competition float64 `json:"competition"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	price, err := h.service.CalculatePrice(c.Request.Context(), req.ProductID, req.SkuID, req.Demand, req.Competition)
	if err != nil {
		h.logger.Error("Failed to calculate price", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to calculate price", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Price calculated successfully", gin.H{"price": price})
}

// ListRules 规则列表
func (h *Handler) ListRules(c *gin.Context) {
	productID, _ := strconv.ParseUint(c.Query("product_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListRules(c.Request.Context(), productID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list rules", "error", err)
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

// ListHistory 历史列表
func (h *Handler) ListHistory(c *gin.Context) {
	productID, _ := strconv.ParseUint(c.Query("product_id"), 10, 64)
	skuID, _ := strconv.ParseUint(c.Query("sku_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListHistory(c.Request.Context(), productID, skuID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list history", "error", err)
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

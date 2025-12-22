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
	service *application.PricingService
	logger  *slog.Logger
}

// NewHandler 处理 HTTP 或 gRPC 请求。
func NewHandler(service *application.PricingService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) CreateRule(c *gin.Context) {
	var req domain.PricingRule
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.CreateRule(c.Request.Context(), &req); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create rule", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create rule", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Rule created successfully", req)
}

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
		h.logger.ErrorContext(c.Request.Context(), "Failed to calculate price", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to calculate price", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Price calculated successfully", gin.H{"price": price})
}

func (h *Handler) ListRules(c *gin.Context) {
	productID, _ := strconv.ParseUint(c.Query("product_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListRules(c.Request.Context(), productID, page, pageSize)
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
	productID, _ := strconv.ParseUint(c.Query("product_id"), 10, 64)
	skuID, _ := strconv.ParseUint(c.Query("sku_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListHistory(c.Request.Context(), productID, skuID, page, pageSize)
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

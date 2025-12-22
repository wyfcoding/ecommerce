package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/flashsale/application"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 结构体定义了FlashSale模块的HTTP处理层。
type Handler struct {
	app    *application.FlashsaleService
	logger *slog.Logger
}

// NewHandler 创建并返回一个新的 FlashSale HTTP Handler 实例。
func NewHandler(app *application.FlashsaleService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// CreateFlashsale 处理创建秒杀活动的HTTP请求。
func (h *Handler) CreateFlashsale(c *gin.Context) {
	var req struct {
		Name          string    `json:"name" binding:"required"`
		ProductID     uint64    `json:"product_id" binding:"required"`
		SkuID         uint64    `json:"sku_id" binding:"required"`
		OriginalPrice int64     `json:"original_price" binding:"required"`
		FlashPrice    int64     `json:"flash_price" binding:"required"`
		TotalStock    int32     `json:"total_stock" binding:"required"`
		LimitPerUser  int32     `json:"limit_per_user" binding:"required"`
		StartTime     time.Time `json:"start_time" binding:"required"`
		EndTime       time.Time `json:"end_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	flashsale, err := h.app.CreateFlashsale(c.Request.Context(), req.Name, req.ProductID, req.SkuID, req.OriginalPrice, req.FlashPrice, req.TotalStock, req.LimitPerUser, req.StartTime, req.EndTime)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create flashsale", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create flashsale", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Flashsale created successfully", flashsale)
}

// GetFlashsale 处理获取秒杀活动详情的HTTP请求。
func (h *Handler) GetFlashsale(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	flashsale, err := h.app.GetFlashsale(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get flashsale", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get flashsale", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Flashsale retrieved successfully", flashsale)
}

// ListFlashsales 处理获取秒杀活动列表的HTTP请求。
func (h *Handler) ListFlashsales(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var status *domain.FlashsaleStatus
	if s := c.Query("status"); s != "" {
		val, _ := strconv.Atoi(s)
		st := domain.FlashsaleStatus(val)
		status = &st
	}

	list, total, err := h.app.ListFlashsales(c.Request.Context(), status, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list flashsales", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list flashsales", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Flashsales listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// PlaceOrder 处理用户下单参与秒杀活动的HTTP请求。
func (h *Handler) PlaceOrder(c *gin.Context) {
	var req struct {
		UserID      uint64 `json:"user_id" binding:"required"`
		FlashsaleID uint64 `json:"flashsale_id" binding:"required"`
		Quantity    int32  `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	order, err := h.app.PlaceOrder(c.Request.Context(), req.UserID, req.FlashsaleID, req.Quantity)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to place order", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to place order", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Order placed successfully", order)
}

// RegisterRoutes 在给定的Gin路由组中注册FlashSale模块的HTTP路由。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/flashsales")
	{
		group.POST("", h.CreateFlashsale)
		group.GET("/:id", h.GetFlashsale)
		group.GET("", h.ListFlashsales)
		group.POST("/orders", h.PlaceOrder)
	}
}

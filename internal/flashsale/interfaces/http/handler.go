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

// CreateFlashsale 处理创建秒杀活动的 HTTP 请求。
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
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	flashsale, err := h.app.CreateFlashsale(c.Request.Context(), req.Name, req.ProductID, req.SkuID, req.OriginalPrice, req.FlashPrice, req.TotalStock, req.LimitPerUser, req.StartTime, req.EndTime)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to create flashsale", "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "flashsale created successfully", flashsale)
}

// GetFlashsale 获取特定秒杀活动的详情信息。
func (h *Handler) GetFlashsale(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id format", "")
		return
	}

	flashsale, err := h.app.GetFlashsale(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get flashsale detail", "flashsale_id", id, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, flashsale)
}

// ListFlashsales 分页获取秒杀活动列表，支持按状态过滤。
func (h *Handler) ListFlashsales(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	var status *domain.FlashsaleStatus
	if s := c.Query("status"); s != "" {
		val, err := strconv.Atoi(s)
		if err == nil {
			st := domain.FlashsaleStatus(val)
			status = &st
		}
	}

	list, total, err := h.app.ListFlashsales(c.Request.Context(), status, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list flashsales", "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, list, total, int32(page), int32(pageSize))
}

// PlaceOrder 处理秒杀抢购下单请求。
func (h *Handler) PlaceOrder(c *gin.Context) {
	var req struct {
		UserID      uint64 `json:"user_id" binding:"required"`
		FlashsaleID uint64 `json:"flashsale_id" binding:"required"`
		Quantity    int32  `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	order, err := h.app.PlaceOrder(c.Request.Context(), req.UserID, req.FlashsaleID, req.Quantity)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to place flashsale order", "user_id", req.UserID, "flashsale_id", req.FlashsaleID, "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "order accepted", order)
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

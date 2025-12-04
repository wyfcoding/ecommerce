package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。
	"time"     // 导入时间包，用于时间解析。

	"github.com/wyfcoding/ecommerce/internal/flashsale/application"   // 导入秒杀模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain/entity" // 导入秒杀模块的领域实体。
	"github.com/wyfcoding/ecommerce/pkg/response"                     // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了FlashSale模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.FlashSaleService // 依赖FlashSale应用服务，处理核心业务逻辑。
	logger  *slog.Logger                  // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 FlashSale HTTP Handler 实例。
func NewHandler(service *application.FlashSaleService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateFlashsale 处理创建秒杀活动的HTTP请求。
// Method: POST
// Path: /flashsales
func (h *Handler) CreateFlashsale(c *gin.Context) {
	// 定义请求体结构，用于接收秒杀活动的创建信息。
	var req struct {
		Name          string    `json:"name" binding:"required"`           // 活动名称，必填。
		ProductID     uint64    `json:"product_id" binding:"required"`     // 商品ID，必填。
		SkuID         uint64    `json:"sku_id" binding:"required"`         // SKU ID，必填。
		OriginalPrice int64     `json:"original_price" binding:"required"` // 原价，必填。
		FlashPrice    int64     `json:"flash_price" binding:"required"`    // 秒杀价，必填。
		TotalStock    int32     `json:"total_stock" binding:"required"`    // 总库存，必填。
		LimitPerUser  int32     `json:"limit_per_user" binding:"required"` // 每人限购数量，必填。
		StartTime     time.Time `json:"start_time" binding:"required"`     // 开始时间，必填。
		EndTime       time.Time `json:"end_time" binding:"required"`       // 结束时间，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建秒杀活动。
	flashsale, err := h.service.CreateFlashsale(c.Request.Context(), req.Name, req.ProductID, req.SkuID, req.OriginalPrice, req.FlashPrice, req.TotalStock, req.LimitPerUser, req.StartTime, req.EndTime)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create flashsale", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create flashsale", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Flashsale created successfully", flashsale)
}

// GetFlashsale 处理获取秒杀活动详情的HTTP请求。
// Method: GET
// Path: /flashsales/:id
func (h *Handler) GetFlashsale(c *gin.Context) {
	// 从URL路径中解析秒杀活动ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层获取秒杀活动详情。
	flashsale, err := h.service.GetFlashsale(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get flashsale", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get flashsale", err.Error())
		return
	}

	// 返回成功的响应，包含秒杀活动详情。
	response.SuccessWithStatus(c, http.StatusOK, "Flashsale retrieved successfully", flashsale)
}

// ListFlashsales 处理获取秒杀活动列表的HTTP请求。
// Method: GET
// Path: /flashsales
func (h *Handler) ListFlashsales(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 从查询参数中获取状态，并转换为实体FlashsaleStatus。
	var status *entity.FlashsaleStatus
	if s := c.Query("status"); s != "" {
		val, _ := strconv.Atoi(s)
		st := entity.FlashsaleStatus(val)
		status = &st
	}

	// 调用应用服务层获取秒杀活动列表。
	list, total, err := h.service.ListFlashsales(c.Request.Context(), status, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list flashsales", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list flashsales", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Flashsales listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// PlaceOrder 处理用户下单参与秒杀活动的HTTP请求。
// Method: POST
// Path: /flashsales/orders
func (h *Handler) PlaceOrder(c *gin.Context) {
	// 定义请求体结构，用于接收下单信息。
	var req struct {
		UserID      uint64 `json:"user_id" binding:"required"`      // 用户ID，必填。
		FlashsaleID uint64 `json:"flashsale_id" binding:"required"` // 秒杀活动ID，必填。
		Quantity    int32  `json:"quantity" binding:"required"`     // 购买数量，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层进行下单操作。
	order, err := h.service.PlaceOrder(c.Request.Context(), req.UserID, req.FlashsaleID, req.Quantity)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to place order", "error", err)
		// 根据应用服务返回的错误类型，可以返回更具体的HTTP状态码。
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to place order", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Order placed successfully", order)
}

// RegisterRoutes 在给定的Gin路由组中注册FlashSale模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /flashsales 路由组，用于所有秒杀相关接口。
	group := r.Group("/flashsales")
	{
		group.POST("", h.CreateFlashsale)   // 创建秒杀活动。
		group.GET("/:id", h.GetFlashsale)   // 获取秒杀活动详情。
		group.GET("", h.ListFlashsales)     // 获取秒杀活动列表。
		group.POST("/orders", h.PlaceOrder) // 用户下单参与秒杀。
		// TODO: 补充更新秒杀活动、删除秒杀活动、取消订单等接口。
	}
}

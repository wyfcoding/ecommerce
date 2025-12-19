package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/application"   // 导入动态定价模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/domain/entity" // 导入动态定价模块的领域实体。
	"github.com/wyfcoding/pkg/response"                                     // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了DynamicPricing模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.DynamicPricingService // 依赖DynamicPricing应用服务，处理核心业务逻辑。
	logger  *slog.Logger                       // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 DynamicPricing HTTP Handler 实例。
func NewHandler(service *application.DynamicPricingService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CalculatePrice 处理计算动态价格的HTTP请求。
// Method: POST
// Path: /pricing/calculate
func (h *Handler) CalculatePrice(c *gin.Context) {
	// 定义请求体结构，使用 entity.PricingRequest 结构体直接绑定。
	var req entity.PricingRequest
	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层计算动态价格。
	price, err := h.service.CalculatePrice(c.Request.Context(), &req)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to calculate price", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to calculate price", err.Error())
		return
	}

	// 返回成功的响应，包含计算出的动态价格。
	response.SuccessWithStatus(c, http.StatusOK, "Price calculated successfully", price)
}

// GetLatestPrice 处理获取指定SKU最新动态价格的HTTP请求。
// Method: GET
// Path: /pricing/sku/:sku_id/latest
func (h *Handler) GetLatestPrice(c *gin.Context) {
	// 从URL路径中解析SKU ID。
	skuID, err := strconv.ParseUint(c.Param("sku_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid SKU ID", err.Error())
		return
	}

	// 调用应用服务层获取最新价格。
	price, err := h.service.GetLatestPrice(c.Request.Context(), skuID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get latest price", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get latest price", err.Error())
		return
	}

	// 返回成功的响应，包含最新价格信息。
	response.SuccessWithStatus(c, http.StatusOK, "Latest price retrieved successfully", price)
}

// SaveStrategy 处理保存（创建或更新）定价策略的HTTP请求。
// Method: POST
// Path: /pricing/strategies
func (h *Handler) SaveStrategy(c *gin.Context) {
	// 定义请求体结构，使用 entity.PricingStrategy 结构体直接绑定。
	var strategy entity.PricingStrategy
	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&strategy); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层保存定价策略。
	if err := h.service.SaveStrategy(c.Request.Context(), &strategy); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to save strategy", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to save strategy", err.Error())
		return
	}

	// 返回成功的响应，包含保存后的策略信息。
	response.SuccessWithStatus(c, http.StatusOK, "Strategy saved successfully", strategy)
}

// ListStrategies 处理获取定价策略列表的HTTP请求。
// Method: GET
// Path: /pricing/strategies
func (h *Handler) ListStrategies(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取策略列表。
	list, total, err := h.service.ListStrategies(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list strategies", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list strategies", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Strategies listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册DynamicPricing模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /pricing 路由组，用于所有动态定价相关接口。
	group := r.Group("/pricing")
	{
		group.POST("/calculate", h.CalculatePrice)         // 计算动态价格。
		group.GET("/sku/:sku_id/latest", h.GetLatestPrice) // 获取SKU最新价格。
		group.POST("/strategies", h.SaveStrategy)          // 保存定价策略。
		group.GET("/strategies", h.ListStrategies)         // 获取定价策略列表。
		// TODO: 补充获取策略详情、更新策略、删除策略等接口。
	}
}

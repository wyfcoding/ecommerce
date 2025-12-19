package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/pricing/application"   // 导入定价模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/pricing/domain/entity" // 导入定价模块的领域实体。
	"github.com/wyfcoding/pkg/response"                             // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了Pricing模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.PricingService // 依赖Pricing应用服务，处理核心业务逻辑。
	logger  *slog.Logger                // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Pricing HTTP Handler 实例。
func NewHandler(service *application.PricingService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateRule 处理创建定价规则的HTTP请求。
// Method: POST
// Path: /pricing/rules
func (h *Handler) CreateRule(c *gin.Context) {
	// 定义请求体结构，使用 entity.PricingRule 结构体直接绑定。
	var req entity.PricingRule
	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建规则。
	if err := h.service.CreateRule(c.Request.Context(), &req); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create rule", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create rule", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Rule created successfully", req)
}

// CalculatePrice 处理计算价格的HTTP请求。
// Method: POST
// Path: /pricing/calculate
func (h *Handler) CalculatePrice(c *gin.Context) {
	// 定义请求体结构，用于接收计算价格所需的参数。
	var req struct {
		ProductID   uint64  `json:"product_id" binding:"required"` // 商品ID，必填。
		SkuID       uint64  `json:"sku_id" binding:"required"`     // SKU ID，必填。
		Demand      float64 `json:"demand"`                        // 需求系数，选填。
		Competition float64 `json:"competition"`                   // 竞争系数，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层计算价格。
	price, err := h.service.CalculatePrice(c.Request.Context(), req.ProductID, req.SkuID, req.Demand, req.Competition)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to calculate price", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to calculate price", err.Error())
		return
	}

	// 返回成功的响应，包含计算出的价格。
	response.SuccessWithStatus(c, http.StatusOK, "Price calculated successfully", gin.H{"price": price})
}

// ListRules 处理获取定价规则列表的HTTP请求。
// Method: GET
// Path: /pricing/rules
func (h *Handler) ListRules(c *gin.Context) {
	// 从查询参数中获取商品ID、页码和每页大小，并设置默认值。
	productID, _ := strconv.ParseUint(c.Query("product_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取规则列表。
	list, total, err := h.service.ListRules(c.Request.Context(), productID, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list rules", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list rules", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Rules listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListHistory 处理获取价格历史列表的HTTP请求。
// Method: GET
// Path: /pricing/history
func (h *Handler) ListHistory(c *gin.Context) {
	// 从查询参数中获取商品ID、SKU ID、页码和每页大小，并设置默认值。
	productID, _ := strconv.ParseUint(c.Query("product_id"), 10, 64)
	skuID, _ := strconv.ParseUint(c.Query("sku_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取价格历史列表。
	list, total, err := h.service.ListHistory(c.Request.Context(), productID, skuID, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list history", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list history", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "History listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册Pricing模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /pricing 路由组，用于所有定价相关接口。
	group := r.Group("/pricing")
	{
		// 定价规则接口。
		group.POST("/rules", h.CreateRule) // 创建定价规则。
		group.GET("/rules", h.ListRules)   // 获取定价规则列表。
		// TODO: 补充获取规则详情、更新规则、删除规则、启用/禁用规则等接口。

		// 价格计算接口。
		group.POST("/calculate", h.CalculatePrice) // 计算商品价格。

		// 价格历史接口。
		group.GET("/history", h.ListHistory) // 获取价格历史列表。
		// TODO: 补充记录价格变动等接口。
	}
}

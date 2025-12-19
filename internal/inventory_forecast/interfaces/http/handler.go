package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/application"   // 导入库存预测模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/domain/entity" // 导入库存预测模块的领域实体。
	"github.com/wyfcoding/pkg/response"                                        // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了InventoryForecast模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.InventoryForecastService // 依赖InventoryForecast应用服务，处理核心业务逻辑。
	logger  *slog.Logger                          // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 InventoryForecast HTTP Handler 实例。
func NewHandler(service *application.InventoryForecastService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GenerateForecast 处理生成销售预测的HTTP请求。
// Method: POST
// Path: /inventory-forecast/forecasts
func (h *Handler) GenerateForecast(c *gin.Context) {
	// 定义请求体结构，用于接收SKU ID。
	var req struct {
		SKUID uint64 `json:"sku_id" binding:"required"` // SKU ID，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层生成销售预测。
	forecast, err := h.service.GenerateForecast(c.Request.Context(), req.SKUID)
	if err != nil {
		h.logger.Error("Failed to generate forecast", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to generate forecast", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Forecast generated successfully", forecast)
}

// GetForecast 处理获取销售预测的HTTP请求。
// Method: GET
// Path: /inventory-forecast/forecasts/:sku_id
func (h *Handler) GetForecast(c *gin.Context) {
	// 从URL路径中解析SKU ID。
	skuID, err := strconv.ParseUint(c.Param("sku_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid SKU ID", err.Error())
		return
	}

	// 调用应用服务层获取销售预测。
	forecast, err := h.service.GetForecast(c.Request.Context(), skuID)
	if err != nil {
		h.logger.Error("Failed to get forecast", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get forecast", err.Error())
		return
	}

	// 返回成功的响应，包含销售预测信息。
	response.SuccessWithStatus(c, http.StatusOK, "Forecast retrieved successfully", forecast)
}

// ListWarnings 处理获取库存预警列表的HTTP请求。
// Method: GET
// Path: /inventory-forecast/warnings
func (h *Handler) ListWarnings(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取库存预警列表。
	list, total, err := h.service.ListWarnings(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list warnings", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list warnings", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Warnings listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListSlowMovingItems 处理获取滞销品列表的HTTP请求。
// Method: GET
// Path: /inventory-forecast/slow-moving-items
func (h *Handler) ListSlowMovingItems(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取滞销品列表。
	list, total, err := h.service.ListSlowMovingItems(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list slow moving items", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list slow moving items", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Slow moving items listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListReplenishmentSuggestions 处理获取补货建议列表的HTTP请求。
// Method: GET
// Path: /inventory-forecast/replenishment-suggestions
func (h *Handler) ListReplenishmentSuggestions(c *gin.Context) {
	// 从查询参数中获取优先级、页码和每页大小，并设置默认值。
	priority := c.Query("priority")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取补货建议列表。
	list, total, err := h.service.ListReplenishmentSuggestions(c.Request.Context(), priority, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list replenishment suggestions", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list replenishment suggestions", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Replenishment suggestions listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListStockoutRisks 处理获取缺货风险列表的HTTP请求。
// Method: GET
// Path: /inventory-forecast/stockout-risks
func (h *Handler) ListStockoutRisks(c *gin.Context) {
	// 从查询参数中获取风险等级、页码和每页大小，并设置默认值。
	level := c.Query("level")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取缺货风险列表。
	list, total, err := h.service.ListStockoutRisks(c.Request.Context(), entity.StockoutRiskLevel(level), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list stockout risks", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list stockout risks", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Stockout risks listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册InventoryForecast模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /inventory-forecast 路由组，用于所有库存预测相关接口。
	group := r.Group("/inventory-forecast")
	{
		group.POST("/forecasts", h.GenerateForecast)                            // 生成销售预测。
		group.GET("/forecasts/:sku_id", h.GetForecast)                          // 获取销售预测。
		group.GET("/warnings", h.ListWarnings)                                  // 获取库存预警列表。
		group.GET("/slow-moving-items", h.ListSlowMovingItems)                  // 获取滞销品列表。
		group.GET("/replenishment-suggestions", h.ListReplenishmentSuggestions) // 获取补货建议列表。
		group.GET("/stockout-risks", h.ListStockoutRisks)                       // 获取缺货风险列表。
		// TODO: 补充更新预测、删除预测等接口。
	}
}

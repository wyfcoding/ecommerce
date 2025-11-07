package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/analytics/service"
)

// AnalyticsHandler 负责处理分析查询的 HTTP 请求。
type AnalyticsHandler struct {
	svc    service.AnalyticsService // 业务逻辑服务接口
	logger *zap.Logger
}

// NewAnalyticsHandler 创建一个新的 AnalyticsHandler 实例。
func NewAnalyticsHandler(svc service.AnalyticsService, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc, logger: logger}
}

// RegisterRoutes 在 Gin 引擎上注册所有分析相关的路由。
func (h *AnalyticsHandler) RegisterRoutes(r *gin.Engine) {
	// 这些接口通常是内部使用，也需要权限验证
	group := r.Group("/api/v1/analytics")
	// group.Use(auth.AdminMiddleware(...)) // 认证和权限中间件
	{
		group.GET("/revenue/total", h.GetTotalRevenue)
		group.GET("/sales/category", h.GetSalesByProductCategory)
		group.GET("/sales/brand", h.GetSalesByProductBrand)
		group.GET("/users/activity_count", h.GetUserActivityCount)
		group.GET("/products/top_n_revenue", h.GetTopNProductsByRevenue)
		group.GET("/conversion_rate", h.GetConversionRate)
	}
}

// GetTotalRevenue 处理获取总销售额的请求。
// 它解析时间范围参数，调用 AnalyticsService 获取总销售额并返回 JSON 响应。
func (h *AnalyticsHandler) GetTotalRevenue(c *gin.Context) {
	startTime, endTime, err := parseTimeRange(c)
	if err != nil {
		h.logger.Warn("GetTotalRevenue: invalid time range parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time range parameters: " + err.Error()})
		return
	}

	total, err := h.svc.GetTotalRevenue(c.Request.Context(), startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get total revenue", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get total revenue: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total_revenue": total})
}

// GetSalesByProductCategory 处理获取按商品分类划分的销售额的请求。
// 它解析时间范围参数，调用 AnalyticsService 获取数据并返回 JSON 响应。
func (h *AnalyticsHandler) GetSalesByProductCategory(c *gin.Context) {
	startTime, endTime, err := parseTimeRange(c)
	if err != nil {
		h.logger.Warn("GetSalesByProductCategory: invalid time range parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time range parameters: " + err.Error()})
		return
	}

	sales, err := h.svc.GetSalesByProductCategory(c.Request.Context(), startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get sales by product category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sales by product category: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sales_by_category": sales})
}

// GetSalesByProductBrand 处理获取按商品品牌划分的销售额的请求。
// 它解析时间范围参数，调用 AnalyticsService 获取数据并返回 JSON 响应。
func (h *AnalyticsHandler) GetSalesByProductBrand(c *gin.Context) {
	startTime, endTime, err := parseTimeRange(c)
	if err != nil {
		h.logger.Warn("GetSalesByProductBrand: invalid time range parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time range parameters: " + err.Error()})
		return
	}

	sales, err := h.svc.GetSalesByProductBrand(c.Request.Context(), startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get sales by product brand", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sales by product brand: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sales_by_brand": sales})
}

// GetUserActivityCount 处理获取活跃用户数的请求。
// 它解析时间范围参数，调用 AnalyticsService 获取数据并返回 JSON 响应。
func (h *AnalyticsHandler) GetUserActivityCount(c *gin.Context) {
	startTime, endTime, err := parseTimeRange(c)
	if err != nil {
		h.logger.Warn("GetUserActivityCount: invalid time range parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time range parameters: " + err.Error()})
		return
	}

	count, err := h.svc.GetUserActivityCount(c.Request.Context(), startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get user activity count", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user activity count: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_activity_count": count})
}

// GetTopNProductsByRevenue 处理获取销售额最高的N个商品的请求。
// 它解析时间范围和N值参数，调用 AnalyticsService 获取数据并返回 JSON 响应。
func (h *AnalyticsHandler) GetTopNProductsByRevenue(c *gin.Context) {
	startTime, endTime, err := parseTimeRange(c)
	if err != nil {
		h.logger.Warn("GetTopNProductsByRevenue: invalid time range parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time range parameters: " + err.Error()})
		return
	}

	nStr := c.DefaultQuery("n", "10")
	n, err := strconv.Atoi(nStr)
	if err != nil || n <= 0 {
		h.logger.Warn("GetTopNProductsByRevenue: invalid N parameter", zap.Error(err), zap.String("n_str", nStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'n' parameter, must be a positive integer"})
		return
	}

	products, err := h.svc.GetTopNProductsByRevenue(c.Request.Context(), n, startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get top N products by revenue", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get top N products by revenue: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"top_products": products})
}

// GetConversionRate 处理获取转化率的请求。
// 它解析时间范围参数，调用 AnalyticsService 获取数据并返回 JSON 响应。
func (h *AnalyticsHandler) GetConversionRate(c *gin.Context) {
	startTime, endTime, err := parseTimeRange(c)
	if err != nil {
		h.logger.Warn("GetConversionRate: invalid time range parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time range parameters: " + err.Error()})
		return
	}

	rate, err := h.svc.GetConversionRate(c.Request.Context(), startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get conversion rate", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get conversion rate: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversion_rate": rate})
}

// parseTimeRange 从 Gin 上下文中解析并返回开始时间和结束时间。
// 如果参数不存在或格式错误，则返回 nil 和错误。
func parseTimeRange(c *gin.Context) (*time.Time, *time.Time, error) {
	var startTime *time.Time
	startStr := c.Query("start_time")
	if startStr != "" {
		t, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid start_time format: %w", err)
		}
		startTime = &t
	}

	var endTime *time.Time
	endStr := c.Query("end_time")
	if endStr != "" {
		t, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid end_time format: %w", err)
		}
		endTime = &t
	}

	return startTime, endTime, nil
}
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/analytics/service"
)

// AnalyticsHandler 负责处理分析查询的 HTTP 请求
type AnalyticsHandler struct {
	svc    service.AnalyticsService
	logger *zap.Logger
}

// NewAnalyticsHandler 创建一个新的 AnalyticsHandler 实例
func NewAnalyticsHandler(svc service.AnalyticsService, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc, logger: logger}
}

// RegisterRoutes 在 Gin 引擎上注册所有分析相关的路由
func (h *AnalyticsHandler) RegisterRoutes(r *gin.Engine) {
	// 这些接口通常是内部使用，也需要权限验证
	group := r.Group("/api/v1/analytics")
	// group.Use(auth.AdminMiddleware(...))
	{
		group.GET("/revenue/total", h.GetTotalRevenue)
		// 可以添加更多查询路由
	}
}

// GetTotalRevenue 处理获取总销售额的请求
func (h *AnalyticsHandler) GetTotalRevenue(c *gin.Context) {
	total, err := h.svc.GetTotalRevenue(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get total revenue", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取总销售额失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total_revenue": total})
}

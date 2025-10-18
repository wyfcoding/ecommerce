package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/admin/service"
	// 伪代码: 模拟认证中间件
	// auth "ecommerce/internal/auth/handler"
)

// AdminHandler 负责处理管理后台的 HTTP 请求
type AdminHandler struct {
	svc    service.AdminService
	logger *zap.Logger
}

// NewAdminHandler 创建一个新的 AdminHandler 实例
func NewAdminHandler(svc service.AdminService, logger *zap.Logger) *AdminHandler {
	return &AdminHandler{svc: svc, logger: logger}
}

// RegisterRoutes 在 Gin 引擎上注册所有管理后台相关的路由
func (h *AdminHandler) RegisterRoutes(r *gin.Engine) {
	// 所有管理后台接口都需要管理员权限
	group := r.Group("/api/v1/admin")
	// group.Use(auth.AuthMiddleware(...), auth.AdminMiddleware(...))
	{
		group.GET("/dashboard/statistics", h.GetDashboardStatistics)
		// 此处可以添加更多路由，例如:
		// group.GET("/users", h.ListUsers)
		// group.GET("/orders", h.ListOrders)
		// group.PUT("/orders/:id/status", h.UpdateOrderStatus)
	}
}

// GetDashboardStatistics 处理获取仪表盘统计数据的请求
func (h *AdminHandler) GetDashboardStatistics(c *gin.Context) {
	stats, err := h.svc.GetDashboardStatistics(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get dashboard statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取仪表盘数据失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

/*
// ListUsers 示例
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	users, err := h.svc.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}
*/
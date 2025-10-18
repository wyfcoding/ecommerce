package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/order/service"
)

// OrderHandler 负责处理订单的 HTTP 请求
type OrderHandler struct {
	svc    service.OrderService
	logger *zap.Logger
}

// NewOrderHandler 创建一个新的 OrderHandler 实例
func NewOrderHandler(svc service.OrderService, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{svc: svc, logger: logger}
}

// RegisterRoutes 在 Gin 引擎上注册所有订单相关的路由
func (h *OrderHandler) RegisterRoutes(r *gin.Engine) {
	// 所有订单操作都需要用户认证
	// 此处假设上游已添加认证中间件
	group := r.Group("/api/v1/orders")
	{
		group.POST("", h.CreateOrder)       // 从购物车创建订单
		group.GET("/:id", h.GetOrder)        // 获取单个订单详情
		group.GET("", h.ListOrders)         // 列出当前用户的所有订单
		group.PUT("/:id/cancel", h.CancelOrder) // 取消订单
	}
}

// CreateOrderRequest 定义了创建订单的请求体
type CreateOrderRequest struct {
	ShippingAddress string `json:"shipping_address" binding:"required"`
	ContactPhone    string `json:"contact_phone" binding:"required"`
	Remarks         string `json:"remarks"`
}

// CreateOrder 处理创建订单的请求
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	// 从上下文中获取 userID
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	order, err := h.svc.CreateOrderFromCart(c.Request.Context(), userID, req.ShippingAddress, req.ContactPhone, req.Remarks)
	if err != nil {
		h.logger.Error("Failed to create order", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建订单失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "订单创建成功，请尽快支付", "order": order})
}

// GetOrder 处理获取订单详情的请求
func (h *OrderHandler) GetOrder(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	orderID, err := h.getOrderIDFromParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.svc.GetOrderDetails(c.Request.Context(), userID, orderID)
	if err != nil {
		// 根据错误类型返回不同状态码
		if err.Error() == "订单不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "无权访问该订单" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取订单详情失败: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, order)
}

// ListOrders 处理列出用户订单的请求
func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	orders, total, err := h.svc.ListUserOrders(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取订单列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":     page,
		"pageSize": pageSize,
		"total":    total,
		"orders":   orders,
	})
}

// CancelOrder 处理取消订单的请求
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	orderID, err := h.getOrderIDFromParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.svc.CancelOrder(c.Request.Context(), userID, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消订单失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "订单已取消", "order": order})
}

// --- 辅助函数 ---

// getUserIDFromContext 从 Gin 上下文中安全地获取用户ID
func (h *OrderHandler) getUserIDFromContext(c *gin.Context) (uint, error) {
	val, exists := c.Get("userID")
	if !exists {
		h.logger.Warn("userID not found in context")
		return 0, fmt.Errorf("用户未认证")
	}
	userID, ok := val.(uint)
	if !ok {
		h.logger.Error("userID in context is not of type uint")
		return 0, fmt.Errorf("用户ID格式错误")
	}
	return userID, nil
}

// getOrderIDFromParam 从 URL 参数中安全地获取订单ID
func (h *OrderHandler) getOrderIDFromParam(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("无效的订单ID格式")
	}
	return uint(id), nil
}

package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/order/application" // 导入订单模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/order/domain"      // 导入订单模块的领域实体。
	"github.com/wyfcoding/pkg/response"                         // 导入统一的响应处理工具。

	"log/slog" // 导入结构化日志库。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
)

// Handler 结构体定义了Order模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.Order // 依赖Order应用服务，处理核心业务逻辑。
	logger  *slog.Logger              // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Order HTTP Handler 实例。
func NewHandler(service *application.Order, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateOrder 处理创建订单的HTTP请求。
// HTTP 方法: POST
// 请求路径: /orders
func (h *Handler) CreateOrder(c *gin.Context) {
	// 定义请求体结构，用于接收订单的创建信息。
	var req struct {
		UserID uint64     `json:"user_id" binding:"required"` // 用户ID，必填。
		Items  []struct { // 订单商品项列表。
			ProductID       uint64 `json:"product_id" binding:"required"`    // 商品ID，必填。
			SkuID           uint64 `json:"sku_id" binding:"required"`        // SKU ID，必填。
			ProductName     string `json:"product_name" binding:"required"`  // 商品名称，必填。
			SkuName         string `json:"sku_name" binding:"required"`      // SKU名称，必填。
			ProductImageURL string `json:"product_image_url"`                // 商品图片URL，选填。
			Price           int64  `json:"price" binding:"required"`         // 单价，必填。
			Quantity        int32  `json:"quantity" binding:"required,gt=0"` // 数量，必填且大于0。
		} `json:"items" binding:"required,dive"` // 订单项列表，至少一项，并对每项进行嵌套验证。
		ShippingAddress struct { // 收货地址信息。
			RecipientName   string `json:"recipient_name" binding:"required"`   // 收货人姓名，必填。
			PhoneNumber     string `json:"phone_number" binding:"required"`     // 手机号，必填。
			Province        string `json:"province" binding:"required"`         // 省份，必填。
			City            string `json:"city" binding:"required"`             // 城市，必填。
			District        string `json:"district" binding:"required"`         // 区县，必填。
			DetailedAddress string `json:"detailed_address" binding:"required"` // 详细地址，必填。
			PostalCode      string `json:"postal_code"`                         // 邮政编码，选填。
		} `json:"shipping_address" binding:"required"` // 收货地址，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 将请求中的商品项转换为领域实体所需的 OrderItem 列表。
	var items []*domain.OrderItem
	for _, item := range req.Items {
		items = append(items, &domain.OrderItem{
			ProductID:       item.ProductID,
			SkuID:           item.SkuID,
			ProductName:     item.ProductName,
			SkuName:         item.SkuName,
			ProductImageURL: item.ProductImageURL,
			Price:           item.Price,
			Quantity:        item.Quantity,
		})
	}

	// 将请求中的收货地址转换为领域实体所需的 ShippingAddress 值对象。
	shippingAddr := &domain.ShippingAddress{
		RecipientName:   req.ShippingAddress.RecipientName,
		PhoneNumber:     req.ShippingAddress.PhoneNumber,
		Province:        req.ShippingAddress.Province,
		City:            req.ShippingAddress.City,
		District:        req.ShippingAddress.District,
		DetailedAddress: req.ShippingAddress.DetailedAddress,
		PostalCode:      req.ShippingAddress.PostalCode,
	}

	// 调用应用服务层创建订单。
	order, err := h.service.CreateOrder(c.Request.Context(), req.UserID, items, shippingAddr)
	if err != nil {
		h.logger.Error("Failed to create order", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create order", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Order created successfully", order)
}

// GetOrder 处理获取订单详情的HTTP请求。
// HTTP 方法: GET
// 请求路径: /orders/:id
func (h *Handler) GetOrder(c *gin.Context) {
	// 从URL路径中解析订单ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层获取订单详情。
	order, err := h.service.GetOrder(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get order", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get order", err.Error())
		return
	}
	if order == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "Order not found", "")
		return
	}

	// 返回成功的响应，包含订单详情。
	response.SuccessWithStatus(c, http.StatusOK, "Order retrieved successfully", order)
}

// UpdateStatus 处理更新订单状态的HTTP请求。
// HTTP 方法: POST
// 请求路径: /orders/:id/status
func (h *Handler) UpdateStatus(c *gin.Context) {
	// 从URL路径中解析订单ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 定义请求体结构，用于接收状态更新操作。
	var req struct {
		Action        string `json:"action" binding:"required,oneof=pay ship deliver complete cancel"` // 操作类型，必填，只能是指定值之一。
		PaymentMethod string `json:"payment_method"`                                                   // 支付方式，仅在支付操作时使用。
		Reason        string `json:"reason"`                                                           // 原因，仅在取消操作时使用。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	var opErr error // 操作结果错误。
	ctx := c.Request.Context()
	operator := "User" // 默认操作人为用户。
	// 从Gin上下文中获取用户ID，如果存在，则作为操作人。这通常由认证中间件设置。
	if userID, exists := c.Get("userID"); exists {
		operator = strconv.FormatUint(userID.(uint64), 10)
	}

	// 根据请求中的操作类型调用应用服务层的对应方法。
	switch req.Action {
	case "pay":
		opErr = h.service.PayOrder(ctx, id, req.PaymentMethod)
	case "ship":
		opErr = h.service.ShipOrder(ctx, id, operator)
	case "deliver":
		opErr = h.service.DeliverOrder(ctx, id, operator)
	case "complete":
		opErr = h.service.CompleteOrder(ctx, id, operator)
	case "cancel":
		opErr = h.service.CancelOrder(ctx, id, operator, req.Reason)
	}

	if opErr != nil {
		h.logger.Error("Failed to update order status", "error", opErr)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update order status", opErr.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Order status updated successfully", nil)
}

// ListOrders 处理获取订单列表的HTTP请求。
// HTTP 方法: GET
// 请求路径: /orders
func (h *Handler) ListOrders(c *gin.Context) {
	// 从查询参数中获取用户ID、状态、页码和每页大小，并设置默认值。
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	statusStr := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var status *int
	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil { // 只有当状态字符串能成功转换为int时才设置过滤状态。
			status = &s
		}
	}

	// 调用应用服务层获取订单列表。
	list, total, err := h.service.ListOrders(c.Request.Context(), userID, status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list orders", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list orders", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Orders listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册Order模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /orders 路由组，用于所有订单相关接口。
	group := r.Group("/orders")
	{
		group.POST("", h.CreateOrder)             // 创建订单。
		group.GET("", h.ListOrders)               // 获取订单列表。
		group.GET("/:id", h.GetOrder)             // 获取订单详情。
		group.POST("/:id/status", h.UpdateStatus) // 更新订单状态（支付、发货、送达、完成、取消）。
		// TODO: 补充订单项详情、退款请求、退款批准等接口。
	}
}

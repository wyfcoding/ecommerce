package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/order/application"
	"github.com/wyfcoding/ecommerce/internal/order/domain"
	"github.com/wyfcoding/pkg/pagination"
	"github.com/wyfcoding/pkg/response"
)

// Handler 结构体定义了Order模块的HTTP处理层。
type Handler struct {
	service *application.OrderService
	logger  *slog.Logger
}

// NewHandler 创建并返回一个新的 Order HTTP Handler 实例。
func NewHandler(service *application.OrderService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateOrder 处理创建订单的HTTP请求。
func (h *Handler) CreateOrder(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
		Items  []struct {
			ProductID       uint64 `json:"product_id" binding:"required"`
			SkuID           uint64 `json:"sku_id" binding:"required"`
			ProductName     string `json:"product_name" binding:"required"`
			SkuName         string `json:"sku_name" binding:"required"`
			ProductImageURL string `json:"product_image_url"`
			Price           int64  `json:"price" binding:"required"`
			Quantity        int32  `json:"quantity" binding:"required,gt=0"`
		} `json:"items" binding:"required,dive"`
		ShippingAddress struct {
			RecipientName   string `json:"recipient_name" binding:"required"`
			PhoneNumber     string `json:"phone_number" binding:"required"`
			Province        string `json:"province" binding:"required"`
			City            string `json:"city" binding:"required"`
			District        string `json:"district" binding:"required"`
			DetailedAddress string `json:"detailed_address" binding:"required"`
			PostalCode      string `json:"postal_code"`
		} `json:"shipping_address" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

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

	shippingAddr := &domain.ShippingAddress{
		RecipientName:   req.ShippingAddress.RecipientName,
		PhoneNumber:     req.ShippingAddress.PhoneNumber,
		Province:        req.ShippingAddress.Province,
		City:            req.ShippingAddress.City,
		District:        req.ShippingAddress.District,
		DetailedAddress: req.ShippingAddress.DetailedAddress,
		PostalCode:      req.ShippingAddress.PostalCode,
	}

	order, err := h.service.CreateOrder(c.Request.Context(), req.UserID, items, shippingAddr)
	if err != nil {
		h.logger.Error("Failed to create order", "error", err)
		response.InternalError(c, "failed to create order: "+err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Order created successfully", order)
}

// GetOrder 获取订单详情
func (h *Handler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid order ID format")
		return
	}

	order, err := h.service.GetOrder(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get order", "id", id, "error", err)
		response.InternalError(c, err.Error())
		return
	}
	if order == nil {
		response.NotFound(c, "order not found")
		return
	}

	response.Success(c, order)
}

// UpdateStatus 更新订单状态
func (h *Handler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid order ID format")
		return
	}

	var req struct {
		Action        string `json:"action" binding:"required,oneof=pay ship deliver complete cancel"`
		PaymentMethod string `json:"payment_method"`
		Reason        string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid input: "+err.Error())
		return
	}

	operator := "System"
	if uid, exists := c.Get("user_id"); exists {
		// 增加稳健的类型检查
		switch v := uid.(type) {
		case uint64:
			operator = strconv.FormatUint(v, 10)
		case string:
			operator = v
		}
	}

	var opErr error
	ctx := c.Request.Context()
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
		h.logger.Error("Failed to update order status", "id", id, "action", req.Action, "error", opErr)
		response.InternalError(c, opErr.Error())
		return
	}

	response.Success(c, nil)
}

// ListOrders 处理获取订单列表的HTTP请求。
func (h *Handler) ListOrders(c *gin.Context) {
	// 1. 严格解析用户 ID
	userIDStr := c.Query("user_id")
	var userID uint64
	if userIDStr != "" {
		uid, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "invalid user_id format")
			return
		}
		userID = uid
	}

	// 2. 使用标准分页 Request
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	pageReq := pagination.NewRequest(page, pageSize)

	// 3. 解析状态过滤
	var status *int
	if statusStr := c.Query("status"); statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err != nil {
			response.BadRequest(c, "invalid status format")
			return
		}
		status = &s
	}

	// 4. 调用业务逻辑
	list, total, err := h.service.ListOrders(c.Request.Context(), userID, status, pageReq.Page, pageReq.PageSize)
	if err != nil {
		h.logger.Error("Failed to list orders", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	// 5. 使用泛型分页 Result
	response.Success(c, pagination.NewResult(total, pageReq, list))
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/orders")
	{
		group.POST("", h.CreateOrder)
		group.GET("", h.ListOrders)
		group.GET("/:id", h.GetOrder)
		group.POST("/:id/status", h.UpdateStatus)
	}
}

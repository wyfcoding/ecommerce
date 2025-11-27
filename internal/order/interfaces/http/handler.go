package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/order/application"
	"github.com/wyfcoding/ecommerce/internal/order/domain/entity"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"log/slog"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *application.OrderService
	logger  *slog.Logger
}

func NewHandler(service *application.OrderService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateOrder 创建订单
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
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	var items []*entity.OrderItem
	for _, item := range req.Items {
		items = append(items, &entity.OrderItem{
			ProductID:       item.ProductID,
			SkuID:           item.SkuID,
			ProductName:     item.ProductName,
			SkuName:         item.SkuName,
			ProductImageURL: item.ProductImageURL,
			Price:           item.Price,
			Quantity:        item.Quantity,
		})
	}

	shippingAddr := &entity.ShippingAddress{
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
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create order", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Order created successfully", order)
}

// GetOrder 获取订单
func (h *Handler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

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

	response.SuccessWithStatus(c, http.StatusOK, "Order retrieved successfully", order)
}

// UpdateStatus 更新订单状态 (Pay/Ship/Deliver/Complete/Cancel)
func (h *Handler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		Action        string `json:"action" binding:"required,oneof=pay ship deliver complete cancel"`
		PaymentMethod string `json:"payment_method"`
		Reason        string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	var opErr error
	ctx := c.Request.Context()
	operator := "User"
	if userID, exists := c.Get("userID"); exists {
		operator = strconv.FormatUint(userID.(uint64), 10)
	}

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

	response.SuccessWithStatus(c, http.StatusOK, "Order status updated successfully", nil)
}

// ListOrders 获取订单列表
func (h *Handler) ListOrders(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	statusStr := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var status *int
	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil {
			status = &s
		}
	}

	list, total, err := h.service.ListOrders(c.Request.Context(), userID, status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list orders", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list orders", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Orders listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/orders")
	{
		group.POST("", h.CreateOrder)
		group.GET("", h.ListOrders)
		group.GET("/:id", h.GetOrder)
		group.POST("/:id/status", h.UpdateStatus)
	}
}

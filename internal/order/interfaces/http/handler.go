package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	pb "github.com/wyfcoding/ecommerce/goapi/order/v1"
	"github.com/wyfcoding/ecommerce/internal/order/application"
	"github.com/wyfcoding/ecommerce/internal/order/domain"
	"github.com/wyfcoding/pkg/pagination"
	"github.com/wyfcoding/pkg/response"
)

// Handler 结构体定义了Order模块的HTTP处理层。
type Handler struct {
	cmdService   *application.OrderCommandService
	queryService *application.OrderQueryService
	logger       *slog.Logger
}

// NewHandler 创建并返回一个新的 Order HTTP Handler 实例。
func NewHandler(cmd *application.OrderCommandService, query *application.OrderQueryService, logger *slog.Logger) *Handler {
	return &Handler{
		cmdService:   cmd,
		queryService: query,
		logger:       logger,
	}
}

// CreateOrder 处理创建订单的 HTTP 请求。
func (h *Handler) CreateOrder(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
		Items  []struct {
			ProductID       uint64 `json:"product_id" binding:"required"`
			SkuID           uint64 `json:"sku_id" binding:"required"`
			ProductName     string `json:"product_name"`
			SkuName         string `json:"sku_name"`
			ProductImageURL string `json:"product_image_url"`
			Price           int64  `json:"price" binding:"required"`
			Quantity        int32  `json:"quantity" binding:"required,gt=0"`
		} `json:"items" binding:"required,dive"`
		ShippingAddress struct {
			RecipientName   string  `json:"recipient_name" binding:"required"`
			PhoneNumber     string  `json:"phone_number" binding:"required"`
			Province        string  `json:"province" binding:"required"`
			City            string  `json:"city" binding:"required"`
			District        string  `json:"district" binding:"required"`
			DetailedAddress string  `json:"detailed_address" binding:"required"`
			PostalCode      string  `json:"postal_code"`
			Lat             float64 `json:"lat"`
			Lon             float64 `json:"lon"`
		} `json:"shipping_address" binding:"required"`
		CouponCode    string `json:"coupon_code"`
		Remark        string `json:"remark"`
		PaymentMethod string `json:"payment_method"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	var items []*application.CreateOrderItemCommand
	for _, item := range req.Items {
		items = append(items, &application.CreateOrderItemCommand{
			ProductID: item.ProductID,
			SkuID:     item.SkuID,
			Quantity:  item.Quantity,
			Price:     item.Price,
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
		Lat:             req.ShippingAddress.Lat,
		Lon:             req.ShippingAddress.Lon,
	}

	deviceID := c.GetHeader("X-Device-ID")
	if deviceID == "" {
		deviceID = c.GetHeader("Device-ID")
	}

	cmd := &application.CreateOrderCommand{
		UserID:          req.UserID,
		Items:           items,
		ShippingAddress: shippingAddr,
		CouponCode:      req.CouponCode,
		Remark:          req.Remark,
		PaymentMethod:   req.PaymentMethod,
		ClientIP:        c.ClientIP(),
		DeviceID:        deviceID,
	}

	order, err := h.cmdService.CreateOrder(c.Request.Context(), cmd)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to create order", "user_id", req.UserID, "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "order created successfully", order)
}

// GetOrder 获取订单详情。
func (h *Handler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid order id format", "")
		return
	}

	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)

	order, err := h.queryService.GetOrder(c.Request.Context(), userID, id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get order detail", "order_id", id, "user_id", userID, "error", err)
		response.Error(c, err)
		return
	}
	if order == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "order not found", "")
		return
	}

	response.Success(c, order)
}

// UpdateStatus 处理订单生命周期的各种状态转移动作。
func (h *Handler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid order id format", "")
		return
	}

	var req struct {
		UserID           uint64 `json:"user_id" binding:"required"`
		Action           string `json:"action" binding:"required,oneof=pay ship deliver complete cancel"`
		PaymentMethod    string `json:"payment_method"`
		Amount           int64  `json:"amount"`
		TransactionID    string `json:"transaction_id"`
		TrackingNumber   string `json:"tracking_number"`
		LogisticsCompany string `json:"logistics_company"`
		Reason           string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid transition request", err.Error())
		return
	}

	operator := "System"
	if uid, exists := c.Get("user_id"); exists {
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
		opErr = h.cmdService.PayOrder(ctx, &application.PayOrderCommand{
			UserID:        req.UserID,
			OrderID:       id,
			PaymentMethod: req.PaymentMethod,
			Amount:        req.Amount,
			TransactionID: req.TransactionID,
		})
	case "ship":
		opErr = h.cmdService.ShipOrder(ctx, &application.ShipOrderCommand{
			UserID:           req.UserID,
			OrderID:          id,
			Operator:         operator,
			TrackingNumber:   req.TrackingNumber,
			LogisticsCompany: req.LogisticsCompany,
		})
	case "deliver":
		opErr = h.cmdService.DeliverOrder(ctx, &application.DeliverOrderCommand{
			UserID:           req.UserID,
			OrderID:          id,
			Operator:         operator,
			TrackingNumber:   req.TrackingNumber,
			LogisticsCompany: req.LogisticsCompany,
		})
	case "complete":
		opErr = h.cmdService.CompleteOrder(ctx, &application.CompleteOrderCommand{UserID: req.UserID, OrderID: id, Operator: operator})
	case "cancel":
		opErr = h.cmdService.CancelOrder(ctx, &application.CancelOrderCommand{UserID: req.UserID, OrderID: id, Operator: operator, Reason: req.Reason})
	}

	if opErr != nil {
		h.logger.ErrorContext(ctx, "failed to update order status", "order_id", id, "action", req.Action, "error", opErr)
		response.Error(c, opErr)
		return
	}

	response.Success(c, nil)
}

// UpdateShippingStatus 处理订单物流状态更新。
func (h *Handler) UpdateShippingStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid order id format", "")
		return
	}

	var req struct {
		UserID            uint64 `json:"user_id" binding:"required"`
		NewShippingStatus string `json:"new_shipping_status" binding:"required"`
		TrackingNumber    string `json:"tracking_number"`
		LogisticsCompany  string `json:"logistics_company"`
		Remark            string `json:"remark"`
		Operator          string `json:"operator"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid shipping status request", err.Error())
		return
	}

	status, err := parseShippingStatus(req.NewShippingStatus)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid shipping status", err.Error())
		return
	}

	operator := req.Operator
	if operator == "" {
		operator = "System"
		if uid, exists := c.Get("user_id"); exists {
			switch v := uid.(type) {
			case uint64:
				operator = strconv.FormatUint(v, 10)
			case string:
				operator = v
			}
		}
	}

	var opErr error
	ctx := c.Request.Context()
	switch status {
	case pb.ShippingStatus_SHIPPING_SHIPPED:
		opErr = h.cmdService.ShipOrder(ctx, &application.ShipOrderCommand{
			UserID:           req.UserID,
			OrderID:          id,
			Operator:         operator,
			TrackingNumber:   req.TrackingNumber,
			LogisticsCompany: req.LogisticsCompany,
		})
	case pb.ShippingStatus_SHIPPING_DELIVERED:
		opErr = h.cmdService.DeliverOrder(ctx, &application.DeliverOrderCommand{
			UserID:           req.UserID,
			OrderID:          id,
			Operator:         operator,
			TrackingNumber:   req.TrackingNumber,
			LogisticsCompany: req.LogisticsCompany,
		})
	default:
		opErr = h.cmdService.UpdateShippingStatus(ctx, &application.UpdateShippingStatusCommand{
			UserID:           req.UserID,
			OrderID:          id,
			Operator:         operator,
			NewStatus:        status,
			TrackingNumber:   req.TrackingNumber,
			LogisticsCompany: req.LogisticsCompany,
			Remark:           req.Remark,
		})
	}

	if opErr != nil {
		h.logger.ErrorContext(ctx, "failed to update shipping status", "order_id", id, "status", status, "error", opErr)
		response.Error(c, opErr)
		return
	}

	response.Success(c, nil)
}

// RequestRefund 处理订单退款申请。
func (h *Handler) RequestRefund(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid order id format", "")
		return
	}

	var req struct {
		UserID       uint64 `json:"user_id" binding:"required"`
		RefundAmount int64  `json:"refund_amount"`
		Reason       string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid refund request", err.Error())
		return
	}

	operator := "System"
	if uid, exists := c.Get("user_id"); exists {
		switch v := uid.(type) {
		case uint64:
			operator = strconv.FormatUint(v, 10)
		case string:
			operator = v
		}
	}

	if err := h.cmdService.RequestRefund(c.Request.Context(), &application.RequestRefundCommand{
		UserID:       req.UserID,
		OrderID:      id,
		Operator:     operator,
		RefundAmount: req.RefundAmount,
		Reason:       req.Reason,
	}); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to request refund", "order_id", id, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// ApproveRefund 处理订单退款审核通过。
func (h *Handler) ApproveRefund(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid order id format", "")
		return
	}

	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid refund approve request", err.Error())
		return
	}

	operator := "System"
	if uid, exists := c.Get("user_id"); exists {
		switch v := uid.(type) {
		case uint64:
			operator = strconv.FormatUint(v, 10)
		case string:
			operator = v
		}
	}

	if err := h.cmdService.ApproveRefund(c.Request.Context(), &application.ApproveRefundCommand{
		UserID:   req.UserID,
		OrderID:  id,
		Operator: operator,
	}); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to approve refund", "order_id", id, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// ListOrders 分页获取订单列表，支持按用户 ID 和状态过滤。
func (h *Handler) ListOrders(c *gin.Context) {
	userIDStr := c.Query("user_id")
	var userID uint64
	if userIDStr != "" {
		uid, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user_id format", "")
			return
		}
		userID = uid
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}
	pageReq := pagination.NewRequest(page, pageSize)

	var status *int
	if statusStr := c.Query("status"); statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil {
			status = &s
		}
	}

	var startTime *time.Time
	if v := c.Query("start_time"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.ErrorWithStatus(c, http.StatusBadRequest, "invalid start_time format", "")
			return
		}
		startTime = &t
	}
	var endTime *time.Time
	if v := c.Query("end_time"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.ErrorWithStatus(c, http.StatusBadRequest, "invalid end_time format", "")
			return
		}
		endTime = &t
	}
	sortBy := c.Query("sort_by")

	var list []*domain.Order
	var total int64
	if userID > 0 {
		list, total, err = h.queryService.ListUserOrders(c.Request.Context(), userID, status, pageReq.Offset(), pageReq.Limit(), startTime, endTime, sortBy)
	} else {
		list, total, err = h.queryService.ListOrders(c.Request.Context(), status, pageReq.Offset(), pageReq.Limit(), startTime, endTime, sortBy)
	}
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list orders", "user_id", userID, "error", err)
		response.Error(c, err)
		return
	}

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
		group.POST("/:id/shipping", h.UpdateShippingStatus)
		group.POST("/:id/refund", h.RequestRefund)
		group.POST("/:id/refund/approve", h.ApproveRefund)
	}
}

func parseShippingStatus(input string) (pb.ShippingStatus, error) {
	val := strings.TrimSpace(input)
	if val == "" {
		return pb.ShippingStatus_SHIPPING_STATUS_UNSPECIFIED, fmt.Errorf("shipping status is required")
	}

	upper := strings.ToUpper(val)
	if v, ok := pb.ShippingStatus_value[upper]; ok {
		return pb.ShippingStatus(v), nil
	}

	if i, err := strconv.Atoi(val); err == nil {
		if _, ok := pb.ShippingStatus_name[int32(i)]; ok {
			return pb.ShippingStatus(i), nil
		}
	}

	return pb.ShippingStatus_SHIPPING_STATUS_UNSPECIFIED, fmt.Errorf("unsupported shipping status: %s", input)
}

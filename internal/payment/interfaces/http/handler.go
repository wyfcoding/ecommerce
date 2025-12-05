package http

import (
	"log/slog"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/payment/application"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 处理支付服务的 HTTP 请求。
type Handler struct {
	app    *application.PaymentApplicationService
	logger *slog.Logger
}

// NewHandler 创建一个新的 HTTP 处理器。
func NewHandler(app *application.PaymentApplicationService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// RegisterRoutes 注册支付相关的路由。
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	payments := router.Group("/payments")
	{
		payments.POST("", h.InitiatePayment)
		payments.POST("/callback", h.HandlePaymentCallback)
		payments.GET("/:id", h.GetPaymentStatus)
		payments.POST("/:id/refunds", h.RequestRefund)
	}
}

type initiatePaymentRequest struct {
	OrderID       uint64 `json:"order_id" binding:"required"`
	UserID        uint64 `json:"user_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	PaymentMethod string `json:"payment_method" binding:"required"`
}

// InitiatePayment 处理支付发起请求。
func (h *Handler) InitiatePayment(c *gin.Context) {
	var req initiatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	payment, err := h.app.InitiatePayment(c.Request.Context(), req.OrderID, req.UserID, req.Amount, req.PaymentMethod)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to initiate payment", "error", err)
		response.InternalError(c, "failed to initiate payment: "+err.Error())
		return
	}

	response.Success(c, payment)
}

type paymentCallbackRequest struct {
	PaymentNo     string `json:"payment_no" binding:"required"`
	Success       bool   `json:"success"`
	TransactionID string `json:"transaction_id"`
	ThirdPartyNo  string `json:"third_party_no"`
}

// HandlePaymentCallback 处理来自第三方支付提供商的支付回调。
func (h *Handler) HandlePaymentCallback(c *gin.Context) {
	var req paymentCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	if err := h.app.HandlePaymentCallback(c.Request.Context(), req.PaymentNo, req.Success, req.TransactionID, req.ThirdPartyNo); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to handle payment callback", "error", err)
		response.InternalError(c, "failed to handle payment callback: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// GetPaymentStatus 处理获取支付状态的请求。
func (h *Handler) GetPaymentStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid payment id: "+err.Error())
		return
	}

	payment, err := h.app.GetPaymentStatus(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get payment status", "id", id, "error", err)
		response.InternalError(c, "failed to get payment status: "+err.Error())
		return
	}
	if payment == nil {
		response.NotFound(c, "payment not found")
		return
	}

	response.Success(c, payment)
}

type requestRefundRequest struct {
	Amount int64  `json:"amount" binding:"required,gt=0"`
	Reason string `json:"reason" binding:"required"`
}

// RequestRefund 处理退款请求。
func (h *Handler) RequestRefund(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid payment id: "+err.Error())
		return
	}

	var req requestRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	refund, err := h.app.RequestRefund(c.Request.Context(), id, req.Amount, req.Reason)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to request refund", "id", id, "error", err)
		response.InternalError(c, "failed to request refund: "+err.Error())
		return
	}

	response.Success(c, refund)
}

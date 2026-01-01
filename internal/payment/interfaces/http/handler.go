package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/payment/application"
	"github.com/wyfcoding/pkg/response"
	"github.com/wyfcoding/pkg/utils/ctxutil"
)

// Handler 支付HTTP处理器
type Handler struct {
	app    *application.PaymentService
	logger *slog.Logger
}

// NewHandler 创建HTTP处理器
func NewHandler(app *application.PaymentService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// RegisterRoutes 注册路由
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

// InitiatePayment 发起支付
func (h *Handler) InitiatePayment(c *gin.Context) {
	var req initiatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request body: "+err.Error(), "")
		return
	}

	ctx := ctxutil.WithIP(c.Request.Context(), c.ClientIP())
	payment, gatewayResp, err := h.app.InitiatePayment(ctx, req.OrderID, req.UserID, req.Amount, req.PaymentMethod)
	if err != nil {
		h.logger.ErrorContext(ctx, "initiate payment failed", "order_id", req.OrderID, "user_id", req.UserID, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "initiate payment failed: "+err.Error(), "")
		return
	}

	response.Success(c, gin.H{
		"payment":          payment,
		"gateway_response": gatewayResp,
	})
}

type paymentCallbackRequest struct {
	PaymentNo     string `json:"payment_no" binding:"required"`
	Success       bool   `json:"success"`
	TransactionID string `json:"transaction_id"`
	ThirdPartyNo  string `json:"third_party_no"`
}

// HandlePaymentCallback 处理支付回调
func (h *Handler) HandlePaymentCallback(c *gin.Context) {
	var req paymentCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid callback payload", "")
		return
	}

	callbackData := map[string]string{
		"payment_no":     req.PaymentNo,
		"status":         strconv.FormatBool(req.Success),
		"transaction_id": req.TransactionID,
		"third_party_no": req.ThirdPartyNo,
	}

	if err := h.app.HandlePaymentCallback(c.Request.Context(), req.PaymentNo, req.Success, req.TransactionID, req.ThirdPartyNo, callbackData); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "payment callback processing failed", "payment_no", req.PaymentNo, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "callback processing error", "")
		return
	}

	response.Success(c, nil)
}

// GetPaymentStatus 查询支付状态
func (h *Handler) GetPaymentStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid payment ID format", "")
		return
	}

	payment, err := h.app.GetPaymentStatus(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to query payment status", "id", id, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "failed to get status", "")
		return
	}
	if payment == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "payment record not found", "")
		return
	}

	response.Success(c, payment)
}

type requestRefundRequest struct {
	Amount int64  `json:"amount" binding:"required,gt=0"`
	Reason string `json:"reason" binding:"required"`
}

// RequestRefund 申请退款
func (h *Handler) RequestRefund(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid ID", "")
		return
	}

	var req requestRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid refund request data", "")
		return
	}

	refund, err := h.app.RequestRefund(c.Request.Context(), id, req.Amount, req.Reason)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "refund initiation failed", "id", id, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "refund failed: "+err.Error(), "")
		return
	}

	response.Success(c, refund)
}

package http

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/payment/application"
	"github.com/wyfcoding/pkg/response"
)

// Handler 支付HTTP处理器
type Handler struct {
	app    *application.Payment
	logger *slog.Logger
}

// NewHandler 创建HTTP处理器
func NewHandler(app *application.Payment, logger *slog.Logger) *Handler {
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
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	payment, gatewayResp, err := h.app.InitiatePayment(c.Request.Context(), req.OrderID, req.UserID, req.Amount, req.PaymentMethod)
	if err != nil {
		// Log detailed error but return generic message if sensitive
		h.logger.ErrorContext(c.Request.Context(), "initiate payment failed", "error", err)
		response.InternalError(c, "initiate payment failed: "+err.Error())
		return
	}

	resp := map[string]any{
		"payment":          payment,
		"gateway_response": gatewayResp,
	}

	response.Success(c, resp)
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
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	// 尝试将所有请求字段作为回调数据传递
	callbackData := map[string]string{
		"payment_no":     req.PaymentNo,
		"status":         strconv.FormatBool(req.Success),
		"transaction_id": req.TransactionID,
		"third_party_no": req.ThirdPartyNo,
	}

	if err := h.app.HandlePaymentCallback(c.Request.Context(), req.PaymentNo, req.Success, req.TransactionID, req.ThirdPartyNo, callbackData); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "handle callback failed", "error", err)
		response.InternalError(c, "handle callback failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// GetPaymentStatus 查询支付状态
func (h *Handler) GetPaymentStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid payment id")
		return
	}

	payment, err := h.app.GetPaymentStatus(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "get status failed", "id", id, "error", err)
		response.InternalError(c, "get status failed")
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

// RequestRefund 申请退款
func (h *Handler) RequestRefund(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid payment id")
		return
	}

	var req requestRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid body: "+err.Error())
		return
	}

	refund, err := h.app.RequestRefund(c.Request.Context(), id, req.Amount, req.Reason)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "refund request failed", "id", id, "error", err)
		response.InternalError(c, "refund request failed: "+err.Error())
		return
	}

	response.Success(c, refund)
}

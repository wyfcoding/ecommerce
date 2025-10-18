package handler

import (
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/payment/service"
)

// PaymentHandler 负责处理支付的 HTTP 请求
type PaymentHandler struct {
	svc    service.PaymentService
	logger *zap.Logger
}

// NewPaymentHandler 创建一个新的 PaymentHandler 实例
func NewPaymentHandler(svc service.PaymentService, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{svc: svc, logger: logger}
}

// RegisterRoutes 在 Gin 引擎上注册所有支付相关的路由
func (h *PaymentHandler) RegisterRoutes(r *gin.Engine) {
	group := r.Group("/api/v1/payments")
	{
		// 用户发起支付的端点
		group.POST("/create", h.CreatePayment)

		// 支付网关的 Webhook 回调端点
		// 这些端点通常不需要用户认证
		group.POST("/webhooks/stripe", h.HandleStripeWebhook)
		group.POST("/webhooks/alipay", h.HandleAlipayWebhook)
	}
}

// CreatePaymentRequest 定义了创建支付的请求体
type CreatePaymentRequest struct {
	OrderSN string  `json:"order_sn" binding:"required"`
	Gateway string  `json:"gateway" binding:"required"` // e.g., "stripe", "alipay"
	Amount  float64 `json:"amount" binding:"required,gt=0"`
}

// CreatePayment 处理用户发起的支付请求
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	// 从上下文中获取 userID
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}
	userID := userIDVal.(uint)

	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	paymentCredential, err := h.svc.CreatePayment(c.Request.Context(), userID, req.OrderSN, req.Amount, req.Gateway)
	if err != nil {
		h.logger.Error("Failed to create payment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建支付失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "支付创建成功",
		"payment_credential": paymentCredential, // 这可能是支付URL, 二维码数据等
		"gateway":            req.Gateway,
	})
}

// HandleStripeWebhook 处理来自 Stripe 的回调
func (h *PaymentHandler) HandleStripeWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536) // 64KB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

	payload, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read Stripe webhook body", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "读取回调数据失败"})
		return
	}

	if err := h.svc.HandleWebhook(c.Request.Context(), "stripe", payload, c.Request.Header); err != nil {
		h.logger.Error("Stripe webhook processing failed", zap.Error(err))
		// 根据错误类型，可以返回 400 或 500，以便 Stripe 重试
		c.JSON(http.StatusBadRequest, gin.H{"error": "处理回调失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// HandleAlipayWebhook 处理来自支付宝的回调
func (h *PaymentHandler) HandleAlipayWebhook(c *gin.Context) {
	// 支付宝的回调通常是 POST 表单形式
	if err := c.Request.ParseForm(); err != nil {
		h.logger.Error("Failed to parse Alipay webhook form", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "解析回调参数失败"})
		return
	}
	// 将整个表单作为 payload
	payload := []byte(c.Request.PostForm.Encode())

	if err := h.svc.HandleWebhook(c.Request.Context(), "alipay", payload, c.Request.Header); err != nil {
		h.logger.Error("Alipay webhook processing failed", zap.Error(err))
		// 支付宝要求返回特定的 "success" 或 "failure" 字符串
		c.String(http.StatusInternalServerError, "failure")
		return
	}

	c.String(http.StatusOK, "success")
}
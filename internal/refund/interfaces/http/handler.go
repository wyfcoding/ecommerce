// Package http 退款服务 HTTP 接口
package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/refund/application"
	"github.com/wyfcoding/ecommerce/internal/refund/domain"
)

type Handler struct {
	service *application.RefundService
}

func NewHandler(service *application.RefundService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/refunds")
	{
		g.POST("", h.CreateRefund)
		g.POST("/:id/approve", h.MerchantApprove)
		g.POST("/:id/reject", h.MerchantReject)
		// g.GET("", h.List)
		// g.GET("/:id", h.Get)
	}
}

type CreateRefundRequest struct {
	OrderID       string `json:"order_id" binding:"required"`
	OrderNo       string `json:"order_no" binding:"required"`
	MerchantID    uint64 `json:"merchant_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Reason        string `json:"reason" binding:"required"`
	Description   string `json:"description"`
	Type          int8   `json:"type" binding:"required"`
	PaymentID     string `json:"payment_id" binding:"required"`
	TransactionID string `json:"transaction_id" binding:"required"`
	Items         []struct {
		OrderItemID  string `json:"order_item_id"`
		SkuID        uint64 `json:"sku_id"`
		Quantity     int32  `json:"quantity"`
		RefundAmount int64  `json:"refund_amount"`
	} `json:"items"`
}

func (h *Handler) CreateRefund(c *gin.Context) {
	var req CreateRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 假设从中间件获取 UserID
	userID := uint64(1001) // TODO: GetUserID(c)

	cmd := application.CreateRefundCmd{
		OrderID:       req.OrderID,
		OrderNo:       req.OrderNo,
		UserID:        userID,
		MerchantID:    req.MerchantID,
		Amount:        req.Amount,
		Reason:        req.Reason,
		Description:   req.Description,
		Type:          domain.RefundType(req.Type),
		PaymentID:     req.PaymentID,
		TransactionID: req.TransactionID,
	}

	for _, item := range req.Items {
		cmd.Items = append(cmd.Items, application.CreateRefundItem{
			OrderItemID:  item.OrderItemID,
			SkuID:        item.SkuID,
			Quantity:     item.Quantity,
			RefundAmount: item.RefundAmount,
		})
	}

	refundNo, err := h.service.CreateRefund(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"refund_no": refundNo})
}

type ReviewRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) MerchantApprove(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	// 假设 OperatorID 为当前登录商家用户
	operatorID := uint64(2001)

	if err := h.service.MerchantApprove(c.Request.Context(), id, operatorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) MerchantReject(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req ReviewRequest
	c.ShouldBindJSON(&req)

	operatorID := uint64(2001)

	if err := h.service.MerchantReject(c.Request.Context(), id, operatorID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

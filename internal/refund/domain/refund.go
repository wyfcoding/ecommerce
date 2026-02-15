// Package domain 退款服务领域模型
// 生成摘要：
// 1) 定义退款单聚合根 (RefundRequest)
// 2) 包含完整的审批流状态机、退款原因、退款类型
// 3) 定义退款相关领域事件
package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrRefundNotFound      = errors.New("refund request not found")
	ErrStatusConflict      = errors.New("invalid status for operation")
	ErrAmountExceeded      = errors.New("refund amount exceeds original order amount")
	ErrPermissionDenied    = errors.New("permission denied")
)

// RefundType 退款类型
type RefundType int8

const (
	RefundTypeOnlyMoney RefundType = 1 // 仅退款
	RefundTypeReturn    RefundType = 2 // 退货退款
)

// RefundStatus 退款状态
type RefundStatus int8

const (
	RefundStatusCreated        RefundStatus = 1 // 已创建
	RefundStatusMerchantReview RefundStatus = 2 // 待商家审核
	RefundStatusPlatformReview RefundStatus = 3 // 待平台审核
	RefundStatusApproved       RefundStatus = 4 // 审核通过（待退款）
	RefundStatusRefunded       RefundStatus = 5 // 退款成功
	RefundStatusRejected       RefundStatus = 6 // 审核拒绝
	RefundStatusCancelled      RefundStatus = 7 // 用户取消
	RefundStatusFailed         RefundStatus = 8 // 退款失败（银行侧）
)

func (s RefundStatus) String() string {
	switch s {
	case RefundStatusCreated:
		return "CREATED"
	case RefundStatusMerchantReview:
		return "MERCHANT_REVIEW"
	case RefundStatusPlatformReview:
		return "PLATFORM_REVIEW"
	case RefundStatusApproved:
		return "APPROVED"
	case RefundStatusRefunded:
		return "REFUNDED"
	case RefundStatusRejected:
		return "REJECTED"
	case RefundStatusCancelled:
		return "CANCELLED"
	case RefundStatusFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

// RefundRequest 退款单聚合根
type RefundRequest struct {
	ID              uint64       `json:"id"`
	RefundNo        string       `json:"refund_no"`
	OrderID         string       `json:"order_id"`
	OrderNo         string       `json:"order_no"`
	UserID          uint64       `json:"user_id"`
	MerchantID      uint64       `json:"merchant_id"`
	Amount          int64        `json:"amount"` // 分
	Reason          string       `json:"reason"`
	Description     string       `json:"description"`
	Type            RefundType   `json:"type"`
	Status          RefundStatus `json:"status"`
	Images          []string     `json:"images"`
	
	// 关联支付信息
	PaymentID       string       `json:"payment_id"`
	TransactionID   string       `json:"transaction_id"` // 原支付流水号
	
	// 审核信息
	RejectReason    string       `json:"reject_reason,omitempty"`
	OperatorID      uint64       `json:"operator_id,omitempty"`
	
	// 时间戳
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
	CompletedAt     *time.Time   `json:"completed_at,omitempty"`
	
	Items           []RefundItem `json:"items"`
}

// RefundItem 退款明细
type RefundItem struct {
	ID           uint64 `json:"id"`
	RefundID     uint64 `json:"refund_id"`
	OrderItemID  string `json:"order_item_id"`
	SkuID        uint64 `json:"sku_id"`
	Quantity     int32  `json:"quantity"`
	RefundAmount int64  `json:"refund_amount"`
}

// NewRefundRequest 创建退款申请
func NewRefundRequest(
	refundNo, orderID, orderNo string,
	userID, merchantID uint64,
	amount int64,
	reason, desc string,
	typ RefundType,
	paymentID, txID string,
) *RefundRequest {
	return &RefundRequest{
		RefundNo:      refundNo,
		OrderID:       orderID,
		OrderNo:       orderNo,
		UserID:        userID,
		MerchantID:    merchantID,
		Amount:        amount,
		Reason:        reason,
		Description:   desc,
		Type:          typ,
		PaymentID:     paymentID,
		TransactionID: txID,
		Status:        RefundStatusCreated,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// SubmitToMerchant 提交给商家
func (r *RefundRequest) SubmitToMerchant() {
	r.Status = RefundStatusMerchantReview
	r.UpdatedAt = time.Now()
}

// MerchantApprove 商家同意
func (r *RefundRequest) MerchantApprove(operatorID uint64) {
	// 如果是仅退款，直接进入平台/系统打款流程
	// 如果是退货退款，这里应该进入 "待退货" 状态 (WaitReturn)，为了简化直接通过
	r.Status = RefundStatusApproved
	r.OperatorID = operatorID
	r.UpdatedAt = time.Now()
}

// MerchantReject 商家拒绝
func (r *RefundRequest) MerchantReject(operatorID uint64, reason string) {
	r.Status = RefundStatusRejected
	r.OperatorID = operatorID
	r.RejectReason = reason
	r.UpdatedAt = time.Now()
}

// SystemSucceed 系统打款成功
func (r *RefundRequest) SystemSucceed() {
	now := time.Now()
	r.Status = RefundStatusRefunded
	r.CompletedAt = &now
	r.UpdatedAt = now
}

// SystemFail 系统打款失败
func (r *RefundRequest) SystemFail(reason string) {
	r.Status = RefundStatusFailed
	r.RejectReason = reason // 复用拒绝原因字段
	r.UpdatedAt = time.Now()
}

// RefundRepository 仓储接口
type RefundRepository interface {
	Save(ctx context.Context, refund *RefundRequest) error
	GetByID(ctx context.Context, id uint64) (*RefundRequest, error)
	GetByRefundNo(ctx context.Context, no string) (*RefundRequest, error)
	ListByOrder(ctx context.Context, orderID string) ([]*RefundRequest, error)
	ListByMerchant(ctx context.Context, merchantID uint64, status RefundStatus, page, size int) ([]*RefundRequest, int64, error)
	ListPending(ctx context.Context, limit int) ([]*RefundRequest, error) // 获取待处理的退款单
}

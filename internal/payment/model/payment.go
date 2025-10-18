package model

import (
	"time"
)

// PaymentStatus 定义支付交易状态的枚举。
type PaymentStatus int32

const (
	PaymentStatusUnspecified PaymentStatus = 0 // 未指定
	Pending                  PaymentStatus = 1 // 待支付/处理中
	Success                  PaymentStatus = 2 // 支付成功
	Failed                   PaymentStatus = 3 // 支付失败
	Refunding                PaymentStatus = 4 // 退款中
	Refunded                 PaymentStatus = 5 // 已退款
	Closed                   PaymentStatus = 6 // 交易关闭 (例如超时未支付)
)

// RefundStatus 定义退款交易状态的枚举。
type RefundStatus int32

const (
	RefundStatusUnspecified RefundStatus = 0 // 未指定
	PendingRefund           RefundStatus = 1 // 退款申请中/处理中
	RefundSuccess           RefundStatus = 2 // 退款成功
	RefundFailed            RefundStatus = 3 // 退款失败
	RefundClosed            RefundStatus = 4 // 退款关闭
)

// PaymentTransaction 代表一笔支付交易的详细信息。
type PaymentTransaction struct {
	ID                   uint64        `gorm:"primarykey" json:"id"`                               // 支付交易ID
	TransactionNo        string        `gorm:"type:varchar(64);uniqueIndex;not null" json:"transaction_no"` // 支付交易编号 (内部唯一)
	OrderID              uint64        `gorm:"index;not null" json:"order_id"`                     // 关联的订单ID
	UserID               uint64        `gorm:"index;not null" json:"user_id"`                      // 用户ID
	PaymentMethod        string        `gorm:"type:varchar(50);not null" json:"payment_method"`    // 支付方式
	Amount               int64         `gorm:"type:bigint;not null" json:"amount"`                 // 支付金额 (单位: 分)
	Status               PaymentStatus `gorm:"type:tinyint;not null" json:"status"`                // 支付交易状态
	GatewayTransactionID string        `gorm:"type:varchar(100)" json:"gateway_transaction_id"`    // 支付网关的交易ID
	GatewayResponse      string        `gorm:"type:text" json:"gateway_response"`                  // 支付网关的原始响应 (JSON)
	Currency             string        `gorm:"type:varchar(10);not null;default:'CNY'" json:"currency"` // 货币类型
	CreatedAt            time.Time     `gorm:"autoCreateTime" json:"created_at"`                   // 交易创建时间
	UpdatedAt            time.Time     `gorm:"autoUpdateTime" json:"updated_at"`                   // 最后更新时间
	PaidAt               *time.Time    `json:"paid_at,omitempty"`                                  // 支付成功时间
	ClosedAt             *time.Time    `json:"closed_at,omitempty"`                                // 交易关闭时间
	DeletedAt            *time.Time    `gorm:"index" json:"deleted_at,omitempty"`                  // 软删除时间
}

// RefundTransaction 代表一笔退款交易的详细信息。
type RefundTransaction struct {
	ID                   uint64       `gorm:"primarykey" json:"id"`                               // 退款交易ID
	RefundNo             string       `gorm:"type:varchar(64);uniqueIndex;not null" json:"refund_no"` // 退款交易编号 (内部唯一)
	PaymentTransactionID uint64       `gorm:"index;not null" json:"payment_transaction_id"`       // 关联的支付交易ID
	OrderID              uint64       `gorm:"index;not null" json:"order_id"`                     // 关联的订单ID
	UserID               uint64       `gorm:"index;not null" json:"user_id"`                      // 用户ID
	RefundAmount         int64        `gorm:"type:bigint;not null" json:"refund_amount"`          // 退款金额 (单位: 分)
	Status               RefundStatus `gorm:"type:tinyint;not null" json:"status"`                // 退款交易状态
	GatewayRefundID      string       `gorm:"type:varchar(100)" json:"gateway_refund_id"`         // 支付网关的退款ID
	GatewayResponse      string       `gorm:"type:text" json:"gateway_response"`                  // 支付网关的原始响应 (JSON)
	Reason               string       `gorm:"type:varchar(500)" json:"reason"`                    // 退款原因
	CreatedAt            time.Time    `gorm:"autoCreateTime" json:"created_at"`                   // 退款申请时间
	UpdatedAt            time.Time    `gorm:"autoUpdateTime" json:"updated_at"`                   // 最后更新时间
	RefundedAt           *time.Time   `json:"refunded_at,omitempty"`                              // 退款成功时间
	ClosedAt             *time.Time   `json:"closed_at,omitempty"`                                // 退款关闭时间
	DeletedAt            *time.Time   `gorm:"index" json:"deleted_at,omitempty"`                  // 软删除时间
}
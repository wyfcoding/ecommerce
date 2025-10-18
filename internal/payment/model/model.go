package model

import (
	"time"

	"gorm.io/gorm"
)

// PaymentStatus 定义了支付流水的状态
type PaymentStatus int

const (
	PaymentStatusPending   PaymentStatus = iota + 1 // 1: 待处理/待支付
	PaymentStatusSuccess                          // 2: 支付成功
	PaymentStatusFailed                           // 3: 支付失败
	PaymentStatusExpired                          // 4: 支付超时
)

// RefundStatus 定义了退款流水的状态
type RefundStatus int

const (
	RefundStatusPending   RefundStatus = iota + 1 // 1: 退款中
	RefundStatusSuccess                         // 2: 退款成功
	RefundStatusFailed                          // 3: 退款失败
)

// PaymentTransaction 支付流水模型
// 记录每一笔支付请求的详细信息
type PaymentTransaction struct {
	ID                uint          `gorm:"primarykey" json:"id"`
	TransactionSN     string        `gorm:"type:varchar(100);uniqueIndex;not null" json:"transaction_sn"` // 支付系统内部唯一的流水号
	OrderSN           string        `gorm:"type:varchar(100);index;not null" json:"order_sn"`           // 关联的业务订单号
	UserID            uint          `gorm:"not null;index" json:"user_id"`                               // 支付用户ID
	Amount            float64       `gorm:"type:decimal(10,2);not null" json:"amount"`                   // 支付金额
	Gateway           string        `gorm:"type:varchar(50);not null" json:"gateway"`                    // 使用的支付网关 (e.g., 'stripe', 'alipay')
	GatewaySN         string        `gorm:"type:varchar(100);index" json:"gateway_sn"`                   // 第三方支付网关返回的流水号
	Status            PaymentStatus `gorm:"not null;default:1" json:"status"`                            // 支付状态
	FailureReason     string        `gorm:"type:varchar(255)" json:"failure_reason"`                     // 支付失败原因
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

// RefundTransaction 退款流水模型
// 记录每一笔退款请求
type RefundTransaction struct {
	ID                      uint         `gorm:"primarykey" json:"id"`
	RefundSN                string       `gorm:"type:varchar(100);uniqueIndex;not null" json:"refund_sn"`      // 系统内部唯一的退款流水号
	OriginalTransactionSN   string       `gorm:"type:varchar(100);index;not null" json:"original_transaction_sn"` // 关联的原始支付流水号
	OrderSN                 string       `gorm:"type:varchar(100);index;not null" json:"order_sn"`            // 关联的订单号
	Amount                  float64      `gorm:"type:decimal(10,2);not null" json:"amount"`                    // 退款金额
	Gateway                 string       `gorm:"type:varchar(50);not null" json:"gateway"`                     // 退款渠道
	GatewayRefundSN         string       `gorm:"type:varchar(100);index" json:"gateway_refund_sn"`            // 第三方支付网关返回的退款流水号
	Status                  RefundStatus `gorm:"not null;default:1" json:"status"`                             // 退款状态
	Reason                  string       `gorm:"type:varchar(255)" json:"reason"`                              // 退款原因
	ProcessedAt             *time.Time   `json:"processed_at"`                                                // 退款处理完成时间
	CreatedAt               time.Time    `json:"created_at"`
	UpdatedAt               time.Time    `json:"updated_at"`
}

// TableName 自定义表名
func (PaymentTransaction) TableName() string {
	return "payment_transactions"
}

func (RefundTransaction) TableName() string {
	return "refund_transactions"
}

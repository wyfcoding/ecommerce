package data

import (
	"time"

	"gorm.io/gorm"
)

// PaymentTransaction represents a payment transaction record.
type PaymentTransaction struct {
	gorm.Model
	PaymentID     string     `gorm:"uniqueIndex;not null;comment:支付系统生成的唯一ID" json:"paymentId"`
	OrderID       uint64     `gorm:"index;not null;comment:关联的订单ID" json:"orderId"`
	UserID        uint64     `gorm:"index;not null;comment:用户ID" json:"userId"`
	Amount        uint64     `gorm:"not null;comment:支付金额 (分)" json:"amount"`
	Currency      string     `gorm:"not null;size:10;comment:货币单位" json:"currency"`
	PaymentMethod string     `gorm:"not null;size:50;comment:支付方式" json:"paymentMethod"`
	Status        string     `gorm:"not null;size:20;comment:支付状态 (PENDING, SUCCESS, FAILED)" json:"status"`
	TransactionNo string     `gorm:"size:100;comment:支付网关交易号" json:"transactionNo"` // From payment gateway
	CallbackData  string     `gorm:"type:text;comment:支付回调原始数据" json:"callbackData"`
	PaidAt        *time.Time `gorm:"comment:支付完成时间" json:"paidAt"`
	// Add other fields like refund status, refund amount, etc.
}

func (PaymentTransaction) TableName() string {
	return "payment_transactions"
}

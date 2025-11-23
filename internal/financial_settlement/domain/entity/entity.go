package entity

import (
	"time"

	"gorm.io/gorm"
)

// SettlementStatus 结算状态
type SettlementStatus string

const (
	SettlementStatusPending    SettlementStatus = "pending"
	SettlementStatusApproved   SettlementStatus = "approved"
	SettlementStatusProcessing SettlementStatus = "processing"
	SettlementStatusCompleted  SettlementStatus = "completed"
	SettlementStatusRejected   SettlementStatus = "rejected"
)

// PaymentStatus 支付状态
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted  PaymentStatus = "completed"
	PaymentStatusFailed     PaymentStatus = "failed"
)

// Settlement 结算
type Settlement struct {
	gorm.Model
	SellerID         uint64           `gorm:"not null;index;comment:商家ID" json:"seller_id"`
	Period           string           `gorm:"type:varchar(32);not null;comment:结算周期" json:"period"`
	StartDate        time.Time        `gorm:"comment:开始日期" json:"start_date"`
	EndDate          time.Time        `gorm:"comment:结束日期" json:"end_date"`
	TotalSalesAmount int64            `gorm:"not null;comment:总销售额" json:"total_sales_amount"`
	CommissionAmount int64            `gorm:"not null;comment:佣金金额" json:"commission_amount"`
	RebateAmount     int64            `gorm:"not null;comment:返利金额" json:"rebate_amount"`
	OtherFees        int64            `gorm:"not null;comment:其他费用" json:"other_fees"`
	FinalAmount      int64            `gorm:"not null;comment:最终结算金额" json:"final_amount"`
	Status           SettlementStatus `gorm:"type:varchar(32);default:'pending';comment:状态" json:"status"`
	ApprovedBy       string           `gorm:"type:varchar(64);comment:审核人" json:"approved_by"`
	ApprovedAt       *time.Time       `gorm:"comment:审核时间" json:"approved_at"`
	RejectionReason  string           `gorm:"type:text;comment:拒绝原因" json:"rejection_reason"`
}

// SettlementOrder 结算订单
type SettlementOrder struct {
	gorm.Model
	SettlementID uint64 `gorm:"not null;index;comment:结算ID" json:"settlement_id"`
	OrderID      uint64 `gorm:"not null;index;comment:订单ID" json:"order_id"`
	Amount       int64  `gorm:"not null;comment:金额" json:"amount"`
	LogisticsFee int64  `gorm:"not null;comment:物流费" json:"logistics_fee"`
	ReturnFee    int64  `gorm:"not null;comment:退货费" json:"return_fee"`
	OtherFee     int64  `gorm:"not null;comment:其他费用" json:"other_fee"`
}

// SettlementPayment 结算支付
type SettlementPayment struct {
	gorm.Model
	SettlementID  uint64        `gorm:"not null;index;comment:结算ID" json:"settlement_id"`
	SellerID      uint64        `gorm:"not null;index;comment:商家ID" json:"seller_id"`
	Amount        int64         `gorm:"not null;comment:支付金额" json:"amount"`
	Status        PaymentStatus `gorm:"type:varchar(32);default:'pending';comment:状态" json:"status"`
	TransactionID string        `gorm:"type:varchar(128);comment:交易流水号" json:"transaction_id"`
	CompletedAt   *time.Time    `gorm:"comment:完成时间" json:"completed_at"`
}

// SettlementStatistics 结算统计
type SettlementStatistics struct {
	TotalSettlements int64     `json:"total_settlements"`
	TotalAmount      int64     `json:"total_amount"`
	AverageAmount    float64   `json:"average_amount"`
	CompletedCount   int64     `json:"completed_count"`
	PendingCount     int64     `json:"pending_count"`
	RejectedCount    int64     `json:"rejected_count"`
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
}

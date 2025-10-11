package data

import (
	"time"

	"gorm.io/gorm"
)

// SettlementRecord represents a record of a settlement transaction.
type SettlementRecord struct {
	gorm.Model
	RecordID         string     `gorm:"uniqueIndex;not null;comment:结算记录唯一ID" json:"recordId"`
	OrderID          uint64     `gorm:"index;not null;comment:关联订单ID" json:"orderId"`
	MerchantID       uint64     `gorm:"index;not null;comment:商家ID" json:"merchantId"`
	TotalAmount      uint64     `gorm:"not null;comment:订单总金额 (分)" json:"totalAmount"`
	PlatformFee      uint64     `gorm:"not null;comment:平台佣金 (分)" json:"platformFee"`
	SettlementAmount uint64     `gorm:"not null;comment:结算给商家的金额 (分)" json:"settlementAmount"`
	Status           string     `gorm:"not null;size:50;comment:结算状态 (PENDING, COMPLETED, FAILED)" json:"status"`
	SettledAt        *time.Time `gorm:"comment:结算完成时间" json:"settledAt"`
}

// TableName specifies the table name for SettlementRecord.
func (SettlementRecord) TableName() string {
	return "settlement_records"
}

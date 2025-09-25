package data

import (
	"time"

	"gorm.io/gorm"
)

<<<<<<< HEAD
// SettlementRecord 结算交易记录
=======
// SettlementRecord represents a record of a settlement transaction.
>>>>>>> 04d1270d593e17e866ec0ca4dad1f5d56021f07d
type SettlementRecord struct {
	gorm.Model
	RecordID         string    `gorm:"uniqueIndex;not null;comment:结算记录唯一ID" json:"recordId"`
	OrderID          uint64    `gorm:"index;not null;comment:关联订单ID" json:"orderId"`
	MerchantID       uint64    `gorm:"index;not null;comment:商家ID" json:"merchantId"`
	TotalAmount      uint64    `gorm:"not null;comment:订单总金额 (分)" json:"totalAmount"`
	PlatformFee      uint64    `gorm:"not null;comment:平台佣金 (分)" json:"platformFee"`
	SettlementAmount uint64    `gorm:"not null;comment:结算给商家的金额 (分)" json:"settlementAmount"`
	Status           string    `gorm:"not null;size:50;comment:结算状态 (PENDING, COMPLETED, FAILED)" json:"status"`
	SettledAt        *time.Time `gorm:"comment:结算完成时间" json:"settledAt"`
}

<<<<<<< HEAD
// TableName 指定 SettlementRecord 的表名
=======
// TableName specifies the table name for SettlementRecord.
>>>>>>> 04d1270d593e17e866ec0ca4dad1f5d56021f07d
func (SettlementRecord) TableName() string {
	return "settlement_records"
}

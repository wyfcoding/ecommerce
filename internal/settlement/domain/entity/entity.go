package entity

import (
	"time"

	"gorm.io/gorm"
)

// SettlementStatus 结算状态
type SettlementStatus int8

const (
	SettlementStatusPending    SettlementStatus = 0 // 待结算
	SettlementStatusProcessing SettlementStatus = 1 // 结算中
	SettlementStatusCompleted  SettlementStatus = 2 // 已完成
	SettlementStatusFailed     SettlementStatus = 3 // 失败
)

// SettlementCycle 结算周期
type SettlementCycle string

const (
	SettlementCycleDaily   SettlementCycle = "DAILY"   // 日结
	SettlementCycleWeekly  SettlementCycle = "WEEKLY"  // 周结
	SettlementCycleMonthly SettlementCycle = "MONTHLY" // 月结
)

// Settlement 结算单实体
type Settlement struct {
	gorm.Model
	SettlementNo     string           `gorm:"type:varchar(64);uniqueIndex;not null;comment:结算单号" json:"settlement_no"`
	MerchantID       uint64           `gorm:"index;not null;comment:商户ID" json:"merchant_id"`
	Cycle            SettlementCycle  `gorm:"type:varchar(32);not null;comment:结算周期" json:"cycle"`
	StartDate        time.Time        `gorm:"not null;comment:开始日期" json:"start_date"`
	EndDate          time.Time        `gorm:"not null;comment:结束日期" json:"end_date"`
	OrderCount       int64            `gorm:"not null;default:0;comment:订单数量" json:"order_count"`
	TotalAmount      uint64           `gorm:"not null;default:0;comment:总金额(分)" json:"total_amount"`
	PlatformFee      uint64           `gorm:"not null;default:0;comment:平台手续费(分)" json:"platform_fee"`
	SettlementAmount uint64           `gorm:"not null;default:0;comment:结算金额(分)" json:"settlement_amount"`
	Status           SettlementStatus `gorm:"type:tinyint;not null;default:0;comment:状态" json:"status"`
	SettledAt        *time.Time       `gorm:"comment:结算时间" json:"settled_at"`
	FailReason       string           `gorm:"type:varchar(255);comment:失败原因" json:"fail_reason"`
}

// SettlementDetail 结算明细实体
type SettlementDetail struct {
	gorm.Model
	SettlementID     uint64 `gorm:"index;not null;comment:结算单ID" json:"settlement_id"`
	OrderID          uint64 `gorm:"index;not null;comment:订单ID" json:"order_id"`
	OrderNo          string `gorm:"type:varchar(64);not null;comment:订单号" json:"order_no"`
	OrderAmount      uint64 `gorm:"not null;comment:订单金额(分)" json:"order_amount"`
	PlatformFee      uint64 `gorm:"not null;comment:平台手续费(分)" json:"platform_fee"`
	SettlementAmount uint64 `gorm:"not null;comment:结算金额(分)" json:"settlement_amount"`
}

// MerchantAccount 商户账户实体
type MerchantAccount struct {
	gorm.Model
	MerchantID    uint64  `gorm:"uniqueIndex;not null;comment:商户ID" json:"merchant_id"`
	Balance       uint64  `gorm:"not null;default:0;comment:余额(分)" json:"balance"`
	FrozenBalance uint64  `gorm:"not null;default:0;comment:冻结金额(分)" json:"frozen_balance"`
	TotalIncome   uint64  `gorm:"not null;default:0;comment:总收入(分)" json:"total_income"`
	TotalWithdraw uint64  `gorm:"not null;default:0;comment:总提现(分)" json:"total_withdraw"`
	FeeRate       float64 `gorm:"type:decimal(5,2);not null;default:0;comment:费率(%)" json:"fee_rate"`
}

// AvailableBalance 可用余额
func (a *MerchantAccount) AvailableBalance() uint64 {
	if a.Balance < a.FrozenBalance {
		return 0
	}
	return a.Balance - a.FrozenBalance
}

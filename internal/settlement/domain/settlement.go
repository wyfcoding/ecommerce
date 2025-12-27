package domain

import (
	"time" // 导入时间库。

	"gorm.io/gorm" // 导入GORM库。
)

// SettlementStatus 定义了结算单的生命周期状态。
type SettlementStatus int8

const (
	SettlementStatusPending    SettlementStatus = 0 // 待结算：结算单已创建，等待处理。
	SettlementStatusProcessing SettlementStatus = 1 // 结算中：结算单正在处理。
	SettlementStatusCompleted  SettlementStatus = 2 // 已完成：结算单已完成结算。
	SettlementStatusFailed     SettlementStatus = 3 // 失败：结算处理失败。
)

// SettlementCycle 定义了结算的周期类型。
type SettlementCycle string

const (
	SettlementCycleDaily   SettlementCycle = "DAILY"   // 日结。
	SettlementCycleWeekly  SettlementCycle = "WEEKLY"  // 周结。
	SettlementCycleMonthly SettlementCycle = "MONTHLY" // 月结。
)

// Settlement 实体是结算模块的聚合根。
// 它代表一个商户在特定周期内的结算汇总信息，包含了结算单号、商户ID、周期、金额统计和状态等。
type Settlement struct {
	gorm.Model                        // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	SettlementNo     string           `gorm:"type:varchar(64);uniqueIndex;not null;comment:结算单号" json:"settlement_no"` // 结算单号，唯一索引，不允许为空。
	MerchantID       uint64           `gorm:"index;not null;comment:商户ID" json:"merchant_id"`                          // 关联的商户ID，索引字段。
	Cycle            SettlementCycle  `gorm:"type:varchar(32);not null;comment:结算周期" json:"cycle"`                     // 结算周期类型。
	StartDate        time.Time        `gorm:"not null;comment:开始日期" json:"start_date"`                                 // 结算周期开始日期。
	EndDate          time.Time        `gorm:"not null;comment:结束日期" json:"end_date"`                                   // 结算周期结束日期。
	OrderCount       int64            `gorm:"not null;default:0;comment:订单数量" json:"order_count"`                      // 包含的订单数量。
	TotalAmount      uint64           `gorm:"not null;default:0;comment:总金额(分)" json:"total_amount"`                   // 订单总金额（单位：分）。
	PlatformFee      uint64           `gorm:"not null;default:0;comment:平台手续费(分)" json:"platform_fee"`                 // 平台收取的手续费总额（单位：分）。
	SettlementAmount uint64           `gorm:"not null;default:0;comment:结算金额(分)" json:"settlement_amount"`             // 最终结算给商户的金额（单位：分）。
	Status           SettlementStatus `gorm:"type:tinyint;not null;default:0;comment:状态" json:"status"`                // 结算单状态，默认为待结算。
	SettledAt        *time.Time       `gorm:"comment:结算时间" json:"settled_at"`                                          // 实际完成结算的时间。
	FailReason       string           `gorm:"type:varchar(255);comment:失败原因" json:"fail_reason"`                       // 结算失败的原因。
}

// SettlementDetail 实体代表结算单中的一个订单明细。
type SettlementDetail struct {
	gorm.Model              // 嵌入gorm.Model。
	SettlementID     uint64 `gorm:"index;not null;comment:结算单ID" json:"settlement_id"`     // 关联的结算单ID，索引字段。
	OrderID          uint64 `gorm:"index;not null;comment:订单ID" json:"order_id"`           // 关联的订单ID，索引字段。
	OrderNo          string `gorm:"type:varchar(64);not null;comment:订单号" json:"order_no"` // 关联的订单号。
	OrderAmount      uint64 `gorm:"not null;comment:订单金额(分)" json:"order_amount"`          // 订单总金额（单位：分）。
	PlatformFee      uint64 `gorm:"not null;comment:平台手续费(分)" json:"platform_fee"`         // 该订单产生的平台手续费（单位：分）。
	SettlementAmount uint64 `gorm:"not null;comment:结算金额(分)" json:"settlement_amount"`     // 该订单结算给商户的金额（单位：分）。
}

// MerchantAccount 实体代表商户的账户信息。
// 包含了商户的余额、冻结金额、总收入、总提现和费率等。
type MerchantAccount struct {
	gorm.Model            // 嵌入gorm.Model。
	MerchantID    uint64  `gorm:"uniqueIndex;not null;comment:商户ID" json:"merchant_id"`               // 商户ID，唯一索引，不允许为空。
	Balance       uint64  `gorm:"not null;default:0;comment:余额(分)" json:"balance"`                    // 账户余额（单位：分）。
	FrozenBalance uint64  `gorm:"not null;default:0;comment:冻结金额(分)" json:"frozen_balance"`           // 冻结金额（单位：分）。
	TotalIncome   uint64  `gorm:"not null;default:0;comment:总收入(分)" json:"total_income"`              // 累计总收入（单位：分）。
	TotalWithdraw uint64  `gorm:"not null;default:0;comment:总提现(分)" json:"total_withdraw"`            // 累计总提现金额（单位：分）。
	FeeRate       float64 `gorm:"type:decimal(5,2);not null;default:0;comment:费率(%)" json:"fee_rate"` // 平台向该商户收取的费率（百分比）。
}

// AvailableBalance 计算商户的可用余额。
// 可用余额 = 余额 - 冻结金额。
func (a *MerchantAccount) AvailableBalance() uint64 {
	if a.Balance < a.FrozenBalance {
		return 0 // 如果余额小于冻结金额，则可用余额为0。
	}
	return a.Balance - a.FrozenBalance
}

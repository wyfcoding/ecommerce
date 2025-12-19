package domain

import (
	"time"

	"gorm.io/gorm" // 导入GORM库。
)

// SettlementStatus 定义了结算单的生命周期状态。
type SettlementStatus string

const (
	SettlementStatusPending    SettlementStatus = "pending"    // 待处理：结算单已生成，等待审核。
	SettlementStatusApproved   SettlementStatus = "approved"   // 已批准：结算单已审核通过，可进行支付。
	SettlementStatusProcessing SettlementStatus = "processing" // 处理中：结算支付正在进行中。
	SettlementStatusCompleted  SettlementStatus = "completed"  // 已完成：结算支付已完成。
	SettlementStatusRejected   SettlementStatus = "rejected"   // 已拒绝：结算单审核未通过。
)

// PaymentStatus 定义了结算支付的生命周期状态。
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"    // 待处理：支付已发起，等待银行或支付机构响应。
	PaymentStatusProcessing PaymentStatus = "processing" // 处理中：支付正在处理。
	PaymentStatusCompleted  PaymentStatus = "completed"  // 已完成：支付成功。
	PaymentStatusFailed     PaymentStatus = "failed"     // 失败：支付失败。
)

// Settlement 实体是财务结算模块的聚合根。
// 它代表一个商家与平台之间的结算单，包含了结算周期、各项金额、状态和审核信息。
type Settlement struct {
	gorm.Model                        // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	SellerID         uint64           `gorm:"not null;index;comment:商家ID" json:"seller_id"`                // 关联的商家ID，索引字段。
	Period           string           `gorm:"type:varchar(32);not null;comment:结算周期" json:"period"`        // 结算周期描述，例如“2023-01”表示2023年1月。
	StartDate        time.Time        `gorm:"comment:开始日期" json:"start_date"`                              // 结算周期的开始日期。
	EndDate          time.Time        `gorm:"comment:结束日期" json:"end_date"`                                // 结算周期的结束日期。
	TotalSalesAmount int64            `gorm:"not null;comment:总销售额" json:"total_sales_amount"`             // 结算周期内的总销售额（单位：分）。
	CommissionAmount int64            `gorm:"not null;comment:佣金金额" json:"commission_amount"`              // 平台收取的佣金金额（单位：分）。
	RebateAmount     int64            `gorm:"not null;comment:返利金额" json:"rebate_amount"`                  // 平台给予商家的返利金额（单位：分）。
	OtherFees        int64            `gorm:"not null;comment:其他费用" json:"other_fees"`                     // 其他费用（例如，广告费、罚款等，单位：分）。
	FinalAmount      int64            `gorm:"not null;comment:最终结算金额" json:"final_amount"`                 // 最终应结算给商家的金额（单位：分）。
	Status           SettlementStatus `gorm:"type:varchar(32);default:'pending';comment:状态" json:"status"` // 结算单状态，默认为待处理。
	ApprovedBy       string           `gorm:"type:varchar(64);comment:审核人" json:"approved_by"`             // 审核人（管理员）的名称。
	ApprovedAt       *time.Time       `gorm:"comment:审核时间" json:"approved_at"`                             // 结算单审核通过时间。
	RejectionReason  string           `gorm:"type:text;comment:拒绝原因" json:"rejection_reason"`              // 结算单被拒绝的原因。
}

// SettlementOrder 实体代表结算单中的一笔订单明细。
type SettlementOrder struct {
	gorm.Model          // 嵌入gorm.Model。
	SettlementID uint64 `gorm:"not null;index;comment:结算ID" json:"settlement_id"` // 关联的结算单ID，索引字段。
	OrderID      uint64 `gorm:"not null;index;comment:订单ID" json:"order_id"`      // 关联的订单ID，索引字段。
	Amount       int64  `gorm:"not null;comment:金额" json:"amount"`                // 订单金额（单位：分）。
	LogisticsFee int64  `gorm:"not null;comment:物流费" json:"logistics_fee"`        // 订单产生的物流费用（单位：分）。
	ReturnFee    int64  `gorm:"not null;comment:退货费" json:"return_fee"`           // 订单产生的退货费用（单位：分）。
	OtherFee     int64  `gorm:"not null;comment:其他费用" json:"other_fee"`           // 订单产生的其他费用（单位：分）。
}

// SettlementPayment 实体代表一笔结算支付记录。
type SettlementPayment struct {
	gorm.Model                  // 嵌入gorm.Model。
	SettlementID  uint64        `gorm:"not null;index;comment:结算ID" json:"settlement_id"`            // 关联的结算单ID，索引字段。
	SellerID      uint64        `gorm:"not null;index;comment:商家ID" json:"seller_id"`                // 关联的商家ID，索引字段。
	Amount        int64         `gorm:"not null;comment:支付金额" json:"amount"`                         // 实际支付的金额（单位：分）。
	Status        PaymentStatus `gorm:"type:varchar(32);default:'pending';comment:状态" json:"status"` // 支付状态，默认为待处理。
	TransactionID string        `gorm:"type:varchar(128);comment:交易流水号" json:"transaction_id"`       // 支付平台的交易流水号。
	CompletedAt   *time.Time    `gorm:"comment:完成时间" json:"completed_at"`                            // 支付完成时间。
}

// SettlementStatistics 结构体用于汇总结算统计数据。
// 此实体通常不直接持久化，而是通过查询计算得出。
type SettlementStatistics struct {
	TotalSettlements int64     `json:"total_settlements"` // 总结算单数量。
	TotalAmount      int64     `json:"total_amount"`      // 总结算金额。
	AverageAmount    float64   `json:"average_amount"`    // 平均结算金额。
	CompletedCount   int64     `json:"completed_count"`   // 已完成的结算单数量。
	PendingCount     int64     `json:"pending_count"`     // 待处理的结算单数量。
	RejectedCount    int64     `json:"rejected_count"`    // 已拒绝的结算单数量。
	StartDate        time.Time `json:"start_date"`        // 统计的开始日期。
	EndDate          time.Time `json:"end_date"`          // 统计的结束日期。
}

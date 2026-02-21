// 生成摘要：
// - 从 aftermarket 服务合并到 aftersales 域。
// - 原 aftermarket 为二手回收/以旧换新领域，属于售后逆向物流范畴。
// - 关键实体：RecycleOrder（回收订单聚合根）。
package domain

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// RecycleStatus 回收订单状态。
type RecycleStatus string

const (
	// StatusValuating 估价中：等待质检中心评估商品价值。
	StatusValuating RecycleStatus = "VALUATING"
	// StatusMailed 已寄出：用户已寄出旧商品。
	StatusMailed RecycleStatus = "MAILED"
	// StatusInspected 质检完成：质检中心已评估完毕。
	StatusInspected RecycleStatus = "INSPECTED"
	// StatusFinished 回收完成：尾款已发放至用户。
	StatusFinished RecycleStatus = "FINISHED"
)

// RecycleOrder 二手回收订单聚合根。
// 支持以旧换新、商品估价、寄卖等非标品电商业务。
// 并发控制策略：乐观锁（gorm.Model 自带 UpdatedAt），业务端通过 Version 额外控制。
type RecycleOrder struct {
	gorm.Model
	// OrderNo 回收订单号，全局唯一。
	OrderNo string `gorm:"type:varchar(64);uniqueIndex" json:"order_no"`
	// UserID 发起回收的用户 ID。
	UserID uint64 `gorm:"index" json:"user_id"`
	// ProductName 回收商品名称。
	ProductName string `gorm:"type:varchar(128)" json:"product_name"`
	// EstimatedAmount 初始估价金额。
	EstimatedAmount decimal.Decimal `gorm:"type:decimal(20,4)" json:"estimated_amount"`
	// FinalAmount 质检后最终回收金额。
	FinalAmount decimal.Decimal `gorm:"type:decimal(20,4)" json:"final_amount"`
	// Status 回收订单当前状态。
	Status RecycleStatus `gorm:"type:varchar(16)" json:"status"`
	// QualityReport 质检报告 JSON。
	QualityReport string `gorm:"type:json;comment:质检报告" json:"quality_report"`
	// InspectedAt 质检完成时间。
	InspectedAt *time.Time `json:"inspected_at"`
}

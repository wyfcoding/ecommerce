// 生成摘要：
// - 从 refund 服务合并到 aftersales 域。
// - 退款是售后流程核心子域，支持多种退款路径（原路返回/钱包/线下）。
// - 关键实体：RefundOrder（退款单）。
package domain

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// RefundStatus 退款单状态。
type RefundStatus string

const (
	// RefundPending 退款待处理。
	RefundPending RefundStatus = "PENDING"
	// RefundProcessing 退款处理中。
	RefundProcessing RefundStatus = "PROCESSING"
	// RefundSuccess 退款成功。
	RefundSuccess RefundStatus = "SUCCESS"
	// RefundFailed 退款失败。
	RefundFailed RefundStatus = "FAILED"
	// RefundSuspended 退款挂起，需人工介入。
	RefundSuspended RefundStatus = "SUSPENDED"
)

// RefundType 退款类型/路径。
type RefundType string

const (
	// RefundOriginal 退款原路返回。
	RefundOriginal RefundType = "ORIGINAL"
	// RefundToWallet 退至电子钱包。
	RefundToWallet RefundType = "WALLET"
	// RefundOffline 线下打款（B2B 场景）。
	RefundOffline RefundType = "OFFLINE"
)

// RefundOrder 退款单实体。
type RefundOrder struct {
	// ID 退款单 ID。
	ID string `json:"id"`
	// OrderID 关联订单 ID。
	OrderID string `json:"order_id"`
	// PaymentID 原支付单 ID。
	PaymentID string `json:"payment_id"`
	// Amount 退款金额。
	Amount decimal.Decimal `json:"amount"`
	// Currency 货币代码。
	Currency string `json:"currency"`
	// RefundType 退款路径类型。
	RefundType RefundType `json:"refund_type"`
	// Status 退款状态。
	Status RefundStatus `json:"status"`
	// Reason 退款原因。
	Reason string `json:"reason"`
	// ChannelTxID 支付渠道退款单号。
	ChannelTxID string `json:"channel_tx_id"`
	// ErrorLog 退款失败错误日志。
	ErrorLog string `json:"error_log"`
	// CreatedAt 创建时间。
	CreatedAt time.Time `json:"created_at"`
	// FinishedAt 完成时间。
	FinishedAt *time.Time `json:"finished_at"`
}

// ErrRefundAmountExceed 退款金额超过原支付金额。
var ErrRefundAmountExceed = errors.New("refund amount exceeds original payment")

// RefundRepository 退款仓储接口。
type RefundRepository interface {
	Save(ctx context.Context, refund *RefundOrder) error
	FindByID(ctx context.Context, id string) (*RefundOrder, error)
	ListByOrderID(ctx context.Context, orderID string) ([]*RefundOrder, error)
}

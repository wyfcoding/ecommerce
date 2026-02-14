// 变更说明：新增预售订单逻辑，支持定金+尾款支付模式，包含定金膨胀与支付期限监控。
// 修正：使用正确的 Proto 常量（pb.OrderType_PRE_SALE, pb.OrderStatus_PENDING_BALANCE 等）。
package domain

import (
	"context"
	"fmt"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/order/v1"
)

// PreSaleStatus 预售阶段状态
type PreSaleStatus int

const (
	PreSaleStatusPendingDeposit PreSaleStatus = 1 // 待付定金
	PreSaleStatusPendingBalance PreSaleStatus = 2 // 待付尾款
	PreSaleStatusPaid           PreSaleStatus = 3 // 已付全款
	PreSaleStatusCancelled      PreSaleStatus = 4 // 已取消
)

// PreSaleOrder 预售订单聚合
type PreSaleOrder struct {
	*Order
	DepositAmount    int64         `json:"deposit_amount"`     // 定金金额
	ExpansionAmount  int64         `json:"expansion_amount"`   // 定金膨胀金额（抵扣额）
	BalanceAmount    int64         `json:"balance_amount"`     // 尾款金额
	BalanceStartTime time.Time     `json:"balance_start_time"` // 尾款开始支付时间
	BalanceEndTime   time.Time     `json:"balance_end_time"`   // 尾款截止支付时间
	PreSaleStatus    PreSaleStatus `json:"presale_status"`
	DepositPaidAt    *time.Time    `json:"deposit_paid_at"`
	BalancePaidAt    *time.Time    `json:"balance_paid_at"`
}

// NewPreSaleOrder 创建预售订单
func NewPreSaleOrder(orderNo string, userID uint64, items []*OrderItem, addr *ShippingAddress, deposit, expansion int64, bStart, bEnd time.Time) *PreSaleOrder {
	order := NewOrder(orderNo, userID, pb.OrderType_PRE_SALE, items, addr)

	// 计算尾款：尾款 = 总价 - 定金膨胀额
	balance := max(order.ActualAmount-expansion, 0)

	return &PreSaleOrder{
		Order:            order,
		DepositAmount:    deposit,
		ExpansionAmount:  expansion,
		BalanceAmount:    balance,
		BalanceStartTime: bStart,
		BalanceEndTime:   bEnd,
		PreSaleStatus:    PreSaleStatusPendingDeposit,
	}
}

// PayDeposit 支付定金
func (p *PreSaleOrder) PayDeposit(ctx context.Context, transactionID string) error {
	if p.PreSaleStatus != PreSaleStatusPendingDeposit {
		return fmt.Errorf("invalid presale status for deposit payment: %d", p.PreSaleStatus)
	}

	p.PreSaleStatus = PreSaleStatusPendingBalance
	now := time.Now()
	p.DepositPaidAt = &now
	p.PaymentStatus = pb.PaymentStatus_PROCESSING // 使用 PROCESSING 表示已付定金待尾款
	p.Status = pb.OrderStatus_PENDING_BALANCE     // 设置为待付尾款状态

	p.AddLog("SYSTEM", "DEPOSIT_PAID", "PENDING_PAYMENT", "PARTIALLY_PAID", fmt.Sprintf("Deposit %d paid, tx: %s", p.DepositAmount, transactionID))
	return nil
}

// PayBalance 支付尾款
func (p *PreSaleOrder) PayBalance(ctx context.Context, transactionID string) error {
	if p.PreSaleStatus != PreSaleStatusPendingBalance {
		return fmt.Errorf("invalid presale status for balance payment: %d", p.PreSaleStatus)
	}

	now := time.Now()
	if now.Before(p.BalanceStartTime) {
		return fmt.Errorf("balance payment has not started yet (starts at %s)", p.BalanceStartTime)
	}
	if now.After(p.BalanceEndTime) {
		p.PreSaleStatus = PreSaleStatusCancelled
		return fmt.Errorf("balance payment deadline exceeded")
	}

	p.PreSaleStatus = PreSaleStatusPaid
	p.BalancePaidAt = &now
	p.PaymentStatus = pb.PaymentStatus_SUCCESS
	p.Status = pb.OrderStatus_PAID // 转入已支付（待发货）状态

	p.AddLog("SYSTEM", "BALANCE_PAID", "PARTIALLY_PAID", "SUCCESS", fmt.Sprintf("Balance %d paid, tx: %s", p.BalanceAmount, transactionID))
	return nil
}

// CanPayBalance 是否可付尾款
func (p *PreSaleOrder) CanPayBalance() bool {
	now := time.Now()
	return p.PreSaleStatus == PreSaleStatusPendingBalance &&
		now.After(p.BalanceStartTime) &&
		now.Before(p.BalanceEndTime)
}

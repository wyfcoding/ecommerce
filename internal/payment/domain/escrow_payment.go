// 变更说明：新增担保交易功能，支持资金冻结、确认收货放款、超时自动放款、争议处理。
// 假设：担保交易默认确认收货期限15天，超时自动确认放款。
package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/payment/v1"
	"github.com/wyfcoding/pkg/eventsourcing"
	"github.com/wyfcoding/pkg/idgen"
)

// --- 担保交易状态 ---

// EscrowStatus 担保交易状态
type EscrowStatus int

const (
	EscrowStatusPending  EscrowStatus = 1 // 待支付
	EscrowStatusFrozen   EscrowStatus = 2 // 资金已冻结
	EscrowStatusReleased EscrowStatus = 3 // 已放款给商家
	EscrowStatusRefunded EscrowStatus = 4 // 已退款给买家
	EscrowStatusDisputed EscrowStatus = 5 // 争议中
	EscrowStatusClosed   EscrowStatus = 6 // 已关闭
)

// --- 担保交易 ---

// EscrowPayment 担保交易聚合根
type EscrowPayment struct {
	eventsourcing.AggregateRoot
	ID                uint64              `json:"id"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
	EscrowPaymentNo   string              `json:"escrow_payment_no"` // 担保交易单号
	OrderID           uint64              `json:"order_id"`
	OrderNo           string              `json:"order_no"`
	UserID            uint64              `json:"user_id"`         // 买家ID
	MerchantID        uint64              `json:"merchant_id"`     // 商家ID
	Amount            int64               `json:"amount"`          // 交易金额（分）
	FrozenAmount      int64               `json:"frozen_amount"`   // 冻结金额（分）
	ReleasedAmount    int64               `json:"released_amount"` // 已放款金额（分）
	RefundedAmount    int64               `json:"refunded_amount"` // 已退款金额（分）
	EscrowStatus      EscrowStatus        `json:"escrow_status"`
	PaymentStatus     pb.PaymentStatus    `json:"payment_status"`
	FrozenAt          *time.Time          `json:"frozen_at"`           // 冻结时间
	ConfirmDeadline   *time.Time          `json:"confirm_deadline"`    // 确认收货截止时间
	ReleasedAt        *time.Time          `json:"released_at"`         // 放款时间
	RefundedAt        *time.Time          `json:"refunded_at"`         // 退款时间
	DisputeReason     string              `json:"dispute_reason"`      // 争议原因
	DisputeResolvedAt *time.Time          `json:"dispute_resolved_at"` // 争议解决时间
	Logs              []*EscrowPaymentLog `json:"logs"`                // 操作日志
	PersistenceVer    int64               `json:"version"`

	// 延迟设置
	ConfirmPeriodDays int `json:"confirm_period_days"` // 确认收货期限（天）
}

// EscrowPaymentLog 担保交易操作日志
type EscrowPaymentLog struct {
	ID              uint64    `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	EscrowPaymentID uint64    `json:"escrow_payment_id"`
	Action          string    `json:"action"`
	Operator        string    `json:"operator"`
	OldStatus       string    `json:"old_status"`
	NewStatus       string    `json:"new_status"`
	Amount          int64     `json:"amount"`
	Remark          string    `json:"remark"`
}

// NewEscrowPayment 创建担保交易
func NewEscrowPayment(orderID uint64, orderNo string, userID, merchantID uint64, amount int64, confirmPeriodDays int, idGenerator idgen.Generator) *EscrowPayment {
	if confirmPeriodDays <= 0 {
		confirmPeriodDays = 15 // 默认15天
	}

	escrowNo := fmt.Sprintf("ESC%d", idGenerator.Generate())

	ep := &EscrowPayment{
		OrderID:           orderID,
		OrderNo:           orderNo,
		UserID:            userID,
		MerchantID:        merchantID,
		Amount:            amount,
		FrozenAmount:      0,
		ReleasedAmount:    0,
		RefundedAmount:    0,
		EscrowStatus:      EscrowStatusPending,
		PaymentStatus:     pb.PaymentStatus_PENDING,
		Logs:              make([]*EscrowPaymentLog, 0),
		ConfirmPeriodDays: confirmPeriodDays,
		PersistenceVer:    1,
	}
	ep.SetID(escrowNo)
	ep.EscrowPaymentNo = escrowNo
	ep.AddLog("CREATE", "System", "", EscrowStatusPending.String(), 0, "Escrow payment created")
	return ep
}

// FreezeFunds 冻结买家资金
func (ep *EscrowPayment) FreezeFunds(ctx context.Context, transactionID string) error {
	if ep.EscrowStatus != EscrowStatusPending {
		return errors.New("can only freeze funds in pending status")
	}

	now := time.Now()
	deadline := now.AddDate(0, 0, ep.ConfirmPeriodDays)

	ep.EscrowStatus = EscrowStatusFrozen
	ep.PaymentStatus = pb.PaymentStatus_SUCCESS
	ep.FrozenAmount = ep.Amount
	ep.FrozenAt = &now
	ep.ConfirmDeadline = &deadline

	ep.AddLog("FREEZE", "System", EscrowStatusPending.String(), EscrowStatusFrozen.String(), ep.Amount, fmt.Sprintf("Funds frozen, deadline: %s", deadline.Format("2006-01-02")))
	return nil
}

// ConfirmReceipt 买家确认收货，释放资金给商家
func (ep *EscrowPayment) ConfirmReceipt(ctx context.Context, operator string) error {
	if ep.EscrowStatus != EscrowStatusFrozen {
		return errors.New("can only confirm receipt in frozen status")
	}

	now := time.Now()
	ep.EscrowStatus = EscrowStatusReleased
	ep.ReleasedAmount = ep.FrozenAmount
	ep.FrozenAmount = 0
	ep.ReleasedAt = &now

	ep.AddLog("CONFIRM_RECEIPT", operator, EscrowStatusFrozen.String(), EscrowStatusReleased.String(), ep.ReleasedAmount, "Buyer confirmed receipt, funds released to merchant")
	return nil
}

// AutoConfirm 超时自动确认收货
func (ep *EscrowPayment) AutoConfirm(ctx context.Context) error {
	if ep.EscrowStatus != EscrowStatusFrozen {
		return errors.New("can only auto confirm in frozen status")
	}

	now := time.Now()
	if ep.ConfirmDeadline != nil && now.Before(*ep.ConfirmDeadline) {
		return errors.New("confirm deadline not reached")
	}

	ep.EscrowStatus = EscrowStatusReleased
	ep.ReleasedAmount = ep.FrozenAmount
	ep.FrozenAmount = 0
	ep.ReleasedAt = &now

	ep.AddLog("AUTO_CONFIRM", "System", EscrowStatusFrozen.String(), EscrowStatusReleased.String(), ep.ReleasedAmount, "Auto confirmed due to deadline reached")
	return nil
}

// RequestRefund 买家申请退款
func (ep *EscrowPayment) RequestRefund(ctx context.Context, reason string, operator string) error {
	if ep.EscrowStatus != EscrowStatusFrozen {
		return errors.New("can only request refund in frozen status")
	}

	ep.EscrowStatus = EscrowStatusDisputed
	ep.DisputeReason = reason

	ep.AddLog("REFUND_REQUEST", operator, EscrowStatusFrozen.String(), EscrowStatusDisputed.String(), 0, reason)
	return nil
}

// ApproveRefund 同意退款（商家同意或平台仲裁）
func (ep *EscrowPayment) ApproveRefund(ctx context.Context, refundAmount int64, operator string) error {
	if ep.EscrowStatus != EscrowStatusDisputed && ep.EscrowStatus != EscrowStatusFrozen {
		return errors.New("invalid status for refund")
	}
	if refundAmount > ep.FrozenAmount {
		return fmt.Errorf("refund amount %d exceeds frozen amount %d", refundAmount, ep.FrozenAmount)
	}

	now := time.Now()
	ep.RefundedAmount = refundAmount
	ep.FrozenAmount -= refundAmount
	ep.RefundedAt = &now

	// 如果还有剩余冻结金额，释放给商家
	if ep.FrozenAmount > 0 {
		ep.ReleasedAmount = ep.FrozenAmount
		ep.FrozenAmount = 0
		ep.ReleasedAt = &now
	}

	ep.EscrowStatus = EscrowStatusRefunded
	ep.DisputeResolvedAt = &now

	ep.AddLog("APPROVE_REFUND", operator, EscrowStatusDisputed.String(), EscrowStatusRefunded.String(), refundAmount, fmt.Sprintf("Refund approved: %d", refundAmount))
	return nil
}

// RejectRefund 拒绝退款
func (ep *EscrowPayment) RejectRefund(ctx context.Context, reason string, operator string) error {
	if ep.EscrowStatus != EscrowStatusDisputed {
		return errors.New("can only reject refund in disputed status")
	}

	now := time.Now()
	ep.EscrowStatus = EscrowStatusFrozen // 回到冻结状态继续等待
	ep.DisputeResolvedAt = &now

	ep.AddLog("REJECT_REFUND", operator, EscrowStatusDisputed.String(), EscrowStatusFrozen.String(), 0, reason)
	return nil
}

// PartialRelease 部分放款（分阶段交付场景）
func (ep *EscrowPayment) PartialRelease(ctx context.Context, amount int64, milestone string, operator string) error {
	if ep.EscrowStatus != EscrowStatusFrozen {
		return errors.New("can only partial release in frozen status")
	}
	if amount > ep.FrozenAmount {
		return fmt.Errorf("amount %d exceeds frozen amount %d", amount, ep.FrozenAmount)
	}

	ep.FrozenAmount -= amount
	ep.ReleasedAmount += amount

	if ep.FrozenAmount == 0 {
		now := time.Now()
		ep.EscrowStatus = EscrowStatusReleased
		ep.ReleasedAt = &now
	}

	ep.AddLog("PARTIAL_RELEASE", operator, EscrowStatusFrozen.String(), ep.EscrowStatus.String(), amount, fmt.Sprintf("Milestone: %s", milestone))
	return nil
}

// ExtendDeadline 延长确认期限
func (ep *EscrowPayment) ExtendDeadline(ctx context.Context, extraDays int, operator string) error {
	if ep.EscrowStatus != EscrowStatusFrozen {
		return errors.New("can only extend deadline in frozen status")
	}
	if ep.ConfirmDeadline == nil {
		return errors.New("no deadline set")
	}

	newDeadline := ep.ConfirmDeadline.AddDate(0, 0, extraDays)
	ep.ConfirmDeadline = &newDeadline

	ep.AddLog("EXTEND_DEADLINE", operator, EscrowStatusFrozen.String(), EscrowStatusFrozen.String(), 0, fmt.Sprintf("Extended to: %s", newDeadline.Format("2006-01-02")))
	return nil
}

// Close 关闭担保交易
func (ep *EscrowPayment) Close(ctx context.Context, reason string, operator string) error {
	if ep.EscrowStatus == EscrowStatusReleased || ep.EscrowStatus == EscrowStatusRefunded {
		return errors.New("cannot close completed escrow")
	}

	oldStatus := ep.EscrowStatus
	ep.EscrowStatus = EscrowStatusClosed

	ep.AddLog("CLOSE", operator, oldStatus.String(), EscrowStatusClosed.String(), 0, reason)
	return nil
}

// IsExpired 检查是否已超过确认期限
func (ep *EscrowPayment) IsExpired() bool {
	if ep.ConfirmDeadline == nil {
		return false
	}
	return time.Now().After(*ep.ConfirmDeadline)
}

// AddLog 添加操作日志
func (ep *EscrowPayment) AddLog(action, operator, oldStatus, newStatus string, amount int64, remark string) {
	log := &EscrowPaymentLog{
		Action:    action,
		Operator:  operator,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Amount:    amount,
		Remark:    remark,
	}
	ep.Logs = append(ep.Logs, log)
}

// String 返回状态字符串
func (s EscrowStatus) String() string {
	switch s {
	case EscrowStatusPending:
		return "PENDING"
	case EscrowStatusFrozen:
		return "FROZEN"
	case EscrowStatusReleased:
		return "RELEASED"
	case EscrowStatusRefunded:
		return "REFUNDED"
	case EscrowStatusDisputed:
		return "DISPUTED"
	case EscrowStatusClosed:
		return "CLOSED"
	default:
		return "UNKNOWN"
	}
}

// --- 担保交易仓储接口 ---

// EscrowPaymentRepository 担保交易仓储接口
type EscrowPaymentRepository interface {
	Save(ctx context.Context, payment *EscrowPayment) error
	Update(ctx context.Context, payment *EscrowPayment) error
	FindByID(ctx context.Context, id uint64) (*EscrowPayment, error)
	FindByEscrowPaymentNo(ctx context.Context, escrowPaymentNo string) (*EscrowPayment, error)
	FindByOrderID(ctx context.Context, orderID uint64) (*EscrowPayment, error)
	FindPendingAutoConfirm(ctx context.Context, before time.Time) ([]*EscrowPayment, error) // 查找待自动确认的担保
}

// --- 担保资金服务接口 ---

// EscrowFundService 担保资金服务接口
type EscrowFundService interface {
	// FreezeFunds 冻结买家资金到担保账户
	FreezeFunds(ctx context.Context, userID uint64, amount int64, escrowPaymentNo string) (string, error)
	// ReleaseFunds 释放资金到商家账户
	ReleaseFunds(ctx context.Context, merchantID uint64, amount int64, escrowPaymentNo string) error
	// RefundFunds 退款到买家账户
	RefundFunds(ctx context.Context, userID uint64, amount int64, escrowPaymentNo string) error
	// GetFrozenBalance 获取担保账户冻结余额
	GetFrozenBalance(ctx context.Context, escrowPaymentNo string) (int64, error)
}

// --- 担保交易事件 ---

// EscrowFrozenEvent 资金冻结事件
type EscrowFrozenEvent struct {
	eventsourcing.BaseEvent
	OrderID    uint64    `json:"order_id"`
	UserID     uint64    `json:"user_id"`
	MerchantID uint64    `json:"merchant_id"`
	Amount     int64     `json:"amount"`
	Deadline   time.Time `json:"deadline"`
}

// EscrowReleasedEvent 资金释放事件
type EscrowReleasedEvent struct {
	eventsourcing.BaseEvent
	OrderID    uint64    `json:"order_id"`
	MerchantID uint64    `json:"merchant_id"`
	Amount     int64     `json:"amount"`
	ReleasedAt time.Time `json:"released_at"`
}

// EscrowRefundedEvent 资金退款事件
type EscrowRefundedEvent struct {
	eventsourcing.BaseEvent
	OrderID    uint64    `json:"order_id"`
	UserID     uint64    `json:"user_id"`
	Amount     int64     `json:"amount"`
	RefundedAt time.Time `json:"refunded_at"`
}

// EscrowDisputedEvent 担保争议事件
type EscrowDisputedEvent struct {
	eventsourcing.BaseEvent
	OrderID uint64 `json:"order_id"`
	UserID  uint64 `json:"user_id"`
	Reason  string `json:"reason"`
}

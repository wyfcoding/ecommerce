package domain

import (
	"fmt"
	"time"
)

// PaymentStatus 支付状态
type PaymentStatus int

const (
	PaymentPending   PaymentStatus = 1
	PaymentSuccess   PaymentStatus = 2
	PaymentFailed    PaymentStatus = 3
	PaymentCancelled PaymentStatus = 4
	PaymentRefunding PaymentStatus = 5
	PaymentRefunded  PaymentStatus = 6
)

func (s PaymentStatus) String() string {
	names := map[PaymentStatus]string{
		PaymentPending:   "Pending",
		PaymentSuccess:   "Success",
		PaymentFailed:    "Failed",
		PaymentCancelled: "Cancelled",
		PaymentRefunding: "Refunding",
		PaymentRefunded:  "Refunded",
	}
	return names[s]
}

// Payment 支付聚合根
type Payment struct {
	ID            uint64
	PaymentNo     string
	OrderID       uint64
	OrderNo       string
	UserID        uint64
	Amount        int64
	PaymentMethod string
	Status        PaymentStatus
	TransactionID string
	ThirdPartyNo  string
	CallbackData  string
	FailureReason string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	PaidAt        *time.Time
	CancelledAt   *time.Time
	RefundedAt    *time.Time
	Logs          []*PaymentLog
	Refunds       []*Refund
}

// Refund 退款实体
type Refund struct {
	ID            uint64
	RefundNo      string
	PaymentID     uint64
	PaymentNo     string
	OrderID       uint64
	OrderNo       string
	UserID        uint64
	RefundAmount  int64
	Reason        string
	Status        PaymentStatus
	ThirdPartyNo  string
	FailureReason string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	RefundedAt    *time.Time
}

// PaymentLog 支付日志值对象
type PaymentLog struct {
	ID        uint64
	PaymentID uint64
	Action    string
	OldStatus string
	NewStatus string
	Remark    string
	CreatedAt time.Time
}

// NewPayment 创建支付
func NewPayment(orderID uint64, orderNo string, userID uint64, amount int64, paymentMethod string) *Payment {
	now := time.Now()
	return &Payment{
		PaymentNo:     generatePaymentNo(),
		OrderID:       orderID,
		OrderNo:       orderNo,
		UserID:        userID,
		Amount:        amount,
		PaymentMethod: paymentMethod,
		Status:        PaymentPending,
		CreatedAt:     now,
		UpdatedAt:     now,
		Logs:          []*PaymentLog{},
		Refunds:       []*Refund{},
	}
}

// CanProcess 是否可以处理支付
func (p *Payment) CanProcess() error {
	if p.Status != PaymentPending {
		return fmt.Errorf("payment status must be Pending, current: %s", p.Status.String())
	}
	return nil
}

// Process 处理支付
func (p *Payment) Process(success bool, transactionID, thirdPartyNo string) error {
	if err := p.CanProcess(); err != nil {
		return err
	}

	oldStatus := p.Status
	p.TransactionID = transactionID
	p.ThirdPartyNo = thirdPartyNo
	p.UpdatedAt = time.Now()

	if success {
		p.Status = PaymentSuccess
		now := time.Now()
		p.PaidAt = &now
	} else {
		p.Status = PaymentFailed
		p.FailureReason = "Payment processing failed"
	}

	p.AddLog("Payment Processed", oldStatus.String(), p.Status.String(), "")

	return nil
}

// CanRefund 是否可以退款
func (p *Payment) CanRefund() error {
	if p.Status != PaymentSuccess {
		return fmt.Errorf("payment status must be Success, current: %s", p.Status.String())
	}
	return nil
}

// CreateRefund 创建退款
func (p *Payment) CreateRefund(refundAmount int64, reason string) (*Refund, error) {
	if err := p.CanRefund(); err != nil {
		return nil, err
	}

	if refundAmount <= 0 || refundAmount > p.Amount {
		return nil, fmt.Errorf("invalid refund amount")
	}

	refund := &Refund{
		RefundNo:     generateRefundNo(),
		PaymentID:    p.ID,
		PaymentNo:    p.PaymentNo,
		OrderID:      p.OrderID,
		OrderNo:      p.OrderNo,
		UserID:       p.UserID,
		RefundAmount: refundAmount,
		Reason:       reason,
		Status:       PaymentRefunding,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	p.Refunds = append(p.Refunds, refund)
	p.Status = PaymentRefunding
	p.UpdatedAt = time.Now()

	p.AddLog("Refund Created", "", "", fmt.Sprintf("Refund amount: %d, Reason: %s", refundAmount, reason))

	return refund, nil
}

// ProcessRefund 处理退款
func (p *Payment) ProcessRefund(refundNo string, success bool) error {
	var refund *Refund
	for _, r := range p.Refunds {
		if r.RefundNo == refundNo {
			refund = r
			break
		}
	}

	if refund == nil {
		return fmt.Errorf("refund not found")
	}

	refund.UpdatedAt = time.Now()

	if success {
		refund.Status = PaymentRefunded
		now := time.Now()
		refund.RefundedAt = &now
		p.Status = PaymentRefunded
		p.RefundedAt = &now
	} else {
		refund.Status = PaymentFailed
		refund.FailureReason = "Refund processing failed"
	}

	p.UpdatedAt = time.Now()
	p.AddLog("Refund Processed", "", "", fmt.Sprintf("Refund: %s, Success: %v", refundNo, success))

	return nil
}

// Cancel 取消支付
func (p *Payment) Cancel(reason string) error {
	if p.Status != PaymentPending {
		return fmt.Errorf("payment cannot be cancelled in current status")
	}

	oldStatus := p.Status
	p.Status = PaymentCancelled
	now := time.Now()
	p.CancelledAt = &now
	p.UpdatedAt = now
	p.FailureReason = reason

	p.AddLog("Payment Cancelled", oldStatus.String(), p.Status.String(), reason)

	return nil
}

// AddLog 添加日志
func (p *Payment) AddLog(action, oldStatus, newStatus, remark string) {
	log := &PaymentLog{
		PaymentID: p.ID,
		Action:    action,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Remark:    remark,
		CreatedAt: time.Now(),
	}
	p.Logs = append(p.Logs, log)
}

func generatePaymentNo() string {
	return fmt.Sprintf("P%s", time.Now().Format("20060102150405"))
}

func generateRefundNo() string {
	return fmt.Sprintf("R%s", time.Now().Format("20060102150405"))
}

package domain

import (
	"fmt"
	"time"
)

// PaymentStatus 支付状态枚举
type PaymentStatus int

const (
	PaymentPending   PaymentStatus = 1 // 待支付：支付已创建但未完成
	PaymentSuccess   PaymentStatus = 2 // 支付成功：交易已确认
	PaymentFailed    PaymentStatus = 3 // 支付失败：交易被拒绝或发生错误
	PaymentCancelled PaymentStatus = 4 // 已取消：交易被用户或系统取消
	PaymentRefunding PaymentStatus = 5 // 退款中：部分或全部退款正在处理
	PaymentRefunded  PaymentStatus = 6 // 已退款：全额退款完成
)

// String 返回状态的英文标识
func (s PaymentStatus) String() string {
	switch s {
	case PaymentPending:
		return "Pending"
	case PaymentSuccess:
		return "Success"
	case PaymentFailed:
		return "Failed"
	case PaymentCancelled:
		return "Cancelled"
	case PaymentRefunding:
		return "Refunding"
	case PaymentRefunded:
		return "Refunded"
	default:
		return "Unknown"
	}
}

// Payment 支付聚合根
// 核心实体，负责管理支付生命周期、状态流转及退款记录
type Payment struct {
	ID            uint64        // 内部唯一ID
	PaymentNo     string        // 业务支付单号（全局唯一）
	OrderID       uint64        // 关联业务订单ID
	OrderNo       string        // 关联业务订单号
	UserID        uint64        // 用户ID
	Amount        int64         // 支付金额（单位：分）
	PaymentMethod string        // 支付方式标识（如 "alipay", "wechat"）
	GatewayType   GatewayType   // 实际通过的网关类型
	Status        PaymentStatus // 当前支付状态
	TransactionID string        // 第三方支付网关的交易流水号
	ThirdPartyNo  string        // 第三方支付网关的订单号
	CallbackData  string        // 原始回调数据快照
	FailureReason string        // 失败或取消原因
	CreatedAt     time.Time     // 创建时间
	UpdatedAt     time.Time     // 更新时间
	PaidAt        *time.Time    // 支付成功时间
	CancelledAt   *time.Time    // 取消时间
	RefundedAt    *time.Time    // 退款完成时间
	Logs          []*PaymentLog // 操作日志集合
	Refunds       []*Refund     // 退款记录集合
}

// Refund 退款实体
// 隶属于 Payment 聚合根
type Refund struct {
	ID              uint64        // 内部唯一ID
	RefundNo        string        // 退款单号（全局唯一）
	PaymentID       uint64        // 关联支付ID
	PaymentNo       string        // 关联支付单号
	OrderID         uint64        // 关联订单ID
	OrderNo         string        // 关联订单号
	UserID          uint64        // 用户ID
	RefundAmount    int64         // 退款金额（单位：分）
	Reason          string        // 退款原因
	Status          PaymentStatus // 退款状态
	ThirdPartyNo    string        // 网关退款流水号
	GatewayRefundID string        // 网关返回的退款ID
	FailureReason   string        // 失败原因
	CreatedAt       time.Time     // 创建时间
	UpdatedAt       time.Time     // 更新时间
	RefundedAt      *time.Time    // 退款成功时间
}

// PaymentLog 支付操作日志（值对象）
type PaymentLog struct {
	ID        uint64    // 内部ID
	PaymentID uint64    // 关联支付ID
	Action    string    // 动作名称
	OldStatus string    // 变更前状态
	NewStatus string    // 变更后状态
	Remark    string    // 备注
	CreatedAt time.Time // 记录时间
}

// NewPayment 创建一个新的支付实体。
func NewPayment(orderID uint64, orderNo string, userID uint64, amount int64, paymentMethod string, gatewayType GatewayType) *Payment {
	now := time.Now()
	payment := &Payment{
		PaymentNo:     generatePaymentNo(),
		OrderID:       orderID,
		OrderNo:       orderNo,
		UserID:        userID,
		Amount:        amount,
		PaymentMethod: paymentMethod,
		GatewayType:   gatewayType,
		Status:        PaymentPending,
		CreatedAt:     now,
		UpdatedAt:     now,
		Logs:          make([]*PaymentLog, 0),
		Refunds:       make([]*Refund, 0),
	}
	payment.AddLog("Payment Initiated", "", PaymentPending.String(), fmt.Sprintf("Amount: %d, Gateway: %s", amount, gatewayType))
	return payment
}

// CanProcess 检查当前支付单是否处于可处理状态（待支付）。
func (p *Payment) CanProcess() error {
	if p.Status != PaymentPending {
		return fmt.Errorf("invalid status for processing: %s", p.Status)
	}
	return nil
}

// Process 处理支付回调结果，更新支付单状态及第三方交易信息。
func (p *Payment) Process(success bool, transactionID, thirdPartyNo string) error {
	// 幂等性检查
	if p.Status == PaymentSuccess && p.TransactionID == transactionID {
		return nil
	}

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
		p.AddLog("Payment Success", oldStatus.String(), p.Status.String(), "TransID: "+transactionID)
	} else {
		p.Status = PaymentFailed
		p.FailureReason = "Payment processing reported failure"
		p.AddLog("Payment Failed", oldStatus.String(), p.Status.String(), p.FailureReason)
	}
	return nil
}

// CanRefund 检查当前支付单是否允许发起退款（已成功支付且未完全退款）。
func (p *Payment) CanRefund() error {
	if p.Status != PaymentSuccess && p.Status != PaymentRefunding && p.Status != PaymentRefunded {
		return fmt.Errorf("payment not successful, current status: %s", p.Status)
	}
	if p.Status == PaymentRefunded {
		return fmt.Errorf("payment fully refunded already")
	}
	return nil
}

// CreateRefund 创建一个新的退款申请记录。
func (p *Payment) CreateRefund(refundAmount int64, reason string) (*Refund, error) {
	if err := p.CanRefund(); err != nil {
		return nil, err
	}
	if refundAmount <= 0 || refundAmount > p.Amount {
		return nil, fmt.Errorf("invalid refund amount: %d", refundAmount)
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
	p.Status = PaymentRefunding // 标记主状态为退款中
	p.UpdatedAt = time.Now()
	p.AddLog("Refund Created", "", "", fmt.Sprintf("Amt: %d, Reason: %s", refundAmount, reason))

	return refund, nil
}

// ProcessRefund 处理退款回调结果，更新对应的退款记录和支付单状态。
func (p *Payment) ProcessRefund(refundNo string, success bool, gatewayRefundID string) error {
	var refund *Refund
	for _, r := range p.Refunds {
		if r.RefundNo == refundNo {
			refund = r
			break
		}
	}
	if refund == nil {
		return fmt.Errorf("refund not found: %s", refundNo)
	}

	// 幂等
	if refund.Status == PaymentRefunded && success {
		return nil
	}

	refund.UpdatedAt = time.Now()
	refund.GatewayRefundID = gatewayRefundID

	if success {
		refund.Status = PaymentRefunded
		now := time.Now()
		refund.RefundedAt = &now

		// 简化逻辑：一次退款成功则标记整单退款成功（需根据实际业务调整为部分退款逻辑）
		p.Status = PaymentRefunded
		p.RefundedAt = &now
		p.AddLog("Refund Success", "", PaymentRefunded.String(), "RefundNo: "+refundNo)
	} else {
		refund.Status = PaymentFailed
		refund.FailureReason = "Gateway rejected refund"
		// 退款失败回滚主状态
		p.Status = PaymentSuccess
		p.AddLog("Refund Failed", "", PaymentFailed.String(), "RefundNo: "+refundNo)
	}
	p.UpdatedAt = time.Now()
	return nil
}

// Cancel 取消支付单（仅限待支付状态）。
func (p *Payment) Cancel(reason string) error {
	if p.Status != PaymentPending {
		return fmt.Errorf("cannot cancel from status: %s", p.Status)
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

// AddLog 记录支付操作日志。
func (p *Payment) AddLog(action, oldStatus, newStatus, remark string) {
	p.Logs = append(p.Logs, &PaymentLog{
		PaymentID: p.ID,
		Action:    action,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Remark:    remark,
		CreatedAt: time.Now(),
	})
}

// 辅助函数
func generatePaymentNo() string {
	return fmt.Sprintf("P%s", time.Now().Format("20060102150405"))
}

func generateRefundNo() string {
	return fmt.Sprintf("R%s", time.Now().Format("20060102150405"))
}

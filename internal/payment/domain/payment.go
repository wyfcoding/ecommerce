package domain

import (
	"fmt"  // 导入格式化库，用于错误信息和编号生成。
	"time" // 导入时间库。
)

// PaymentStatus 支付状态
type PaymentStatus int

const (
	PaymentPending   PaymentStatus = 1 // 待处理/待支付：支付已发起但尚未完成。
	PaymentSuccess   PaymentStatus = 2 // 成功：支付已成功完成。
	PaymentFailed    PaymentStatus = 3 // 失败：支付处理失败。
	PaymentCancelled PaymentStatus = 4 // 取消：支付被取消。
	PaymentRefunding PaymentStatus = 5 // 退款中：正在处理退款。
	PaymentRefunded  PaymentStatus = 6 // 已退款：退款已成功完成。
)

// String 返回 PaymentStatus 的字符串表示。
func (s PaymentStatus) String() string {
	names := map[PaymentStatus]string{
		PaymentPending:   "Pending",
		PaymentSuccess:   "Success",
		PaymentFailed:    "Failed",
		PaymentCancelled: "Cancelled",
		PaymentRefunding: "Refunding",
		PaymentRefunded:  "Refunded",
	}
	return names[s] // 根据状态值返回对应的字符串。
}

// Payment 实体是支付模块的聚合根。
// 它代表一个支付单的完整生命周期，包含支付编号、关联订单、用户、金额、状态、交易详情和日志等。
type Payment struct {
	ID            uint64        // 支付ID。
	PaymentNo     string        // 支付单号，唯一标识一笔支付。
	OrderID       uint64        // 关联的订单ID。
	OrderNo       string        // 关联的订单编号。
	UserID        uint64        // 支付用户ID。
	Amount        int64         // 支付金额（单位：分）。
	PaymentMethod string        // 支付方式（例如，支付宝，微信支付，银行卡）。
	Status        PaymentStatus // 当前支付状态。
	TransactionID string        // 第三方支付平台返回的交易流水号。
	ThirdPartyNo  string        // 第三方支付平台返回的支付单号。
	CallbackData  string        // 支付回调的原始数据，可能为JSON字符串。
	FailureReason string        // 支付失败或取消的原因。
	CreatedAt     time.Time     // 支付记录创建时间。
	UpdatedAt     time.Time     // 支付记录最后更新时间。
	PaidAt        *time.Time    // 实际支付成功时间。
	CancelledAt   *time.Time    // 支付取消时间。
	RefundedAt    *time.Time    // 支付全额退款时间。
	Logs          []*PaymentLog // 关联的支付操作日志列表。
	Refunds       []*Refund     // 关联的退款记录列表。
}

// Refund 实体代表一笔退款记录。
// 它是Payment聚合根的一部分。
type Refund struct {
	ID            uint64        // 退款ID。
	RefundNo      string        // 退款单号，唯一标识一笔退款。
	PaymentID     uint64        // 关联的支付ID。
	PaymentNo     string        // 关联的支付单号。
	OrderID       uint64        // 关联的订单ID。
	OrderNo       string        // 关联的订单编号。
	UserID        uint64        // 退款用户ID。
	RefundAmount  int64         // 退款金额（单位：分）。
	Reason        string        // 退款原因。
	Status        PaymentStatus // 退款状态（PaymentRefunding, PaymentRefunded, PaymentFailed）。
	ThirdPartyNo  string        // 第三方支付平台返回的退款流水号。
	FailureReason string        // 退款失败的原因。
	CreatedAt     time.Time     // 退款记录创建时间。
	UpdatedAt     time.Time     // 退款记录最后更新时间。
	RefundedAt    *time.Time    // 实际退款成功时间。
}

// PaymentLog 值对象定义了支付的操作日志记录。
type PaymentLog struct {
	ID        uint64    // 日志ID。
	PaymentID uint64    // 关联的支付ID。
	Action    string    // 操作动作（例如，"Initiated", "Processed", "Cancelled", "Refunded"）。
	OldStatus string    // 操作前的支付状态。
	NewStatus string    // 操作后的支付状态。
	Remark    string    // 备注信息。
	CreatedAt time.Time // 日志创建时间。
}

// NewPayment 创建并返回一个新的 Payment 实体实例。
// orderID: 关联的订单ID。
// orderNo: 关联的订单编号。
// userID: 支付用户ID。
// amount: 支付金额。
// paymentMethod: 支付方式。
func NewPayment(orderID uint64, orderNo string, userID uint64, amount int64, paymentMethod string) *Payment {
	now := time.Now()
	payment := &Payment{
		PaymentNo:     generatePaymentNo(), // 生成唯一的支付单号。
		OrderID:       orderID,
		OrderNo:       orderNo,
		UserID:        userID,
		Amount:        amount,
		PaymentMethod: paymentMethod,
		Status:        PaymentPending, // 初始状态为待处理。
		CreatedAt:     now,
		UpdatedAt:     now,
		Logs:          []*PaymentLog{}, // 初始化日志列表。
		Refunds:       []*Refund{},     // 初始化退款列表。
	}
	payment.AddLog("Payment Initiated", "", PaymentPending.String(), fmt.Sprintf("Amount: %d, Method: %s", amount, paymentMethod))
	return payment
}

// CanProcess 检查支付是否可以进行处理（从待处理状态）。
func (p *Payment) CanProcess() error {
	if p.Status != PaymentPending {
		return fmt.Errorf("payment status must be Pending, current: %s", p.Status.String())
	}
	return nil
}

// Process 处理支付回调的结果。
// success: 支付是否成功。
// transactionID: 第三方交易ID。
// thirdPartyNo: 第三方支付流水号。
func (p *Payment) Process(success bool, transactionID, thirdPartyNo string) error {
	if err := p.CanProcess(); err != nil {
		return err
	}

	oldStatus := p.Status
	p.TransactionID = transactionID
	p.ThirdPartyNo = thirdPartyNo
	p.UpdatedAt = time.Now()

	if success {
		p.Status = PaymentSuccess // 支付成功。
		now := time.Now()
		p.PaidAt = &now // 记录支付成功时间。
		p.AddLog("Payment Success", oldStatus.String(), p.Status.String(), fmt.Sprintf("TransactionID: %s", transactionID))
	} else {
		p.Status = PaymentFailed // 支付失败。
		p.FailureReason = "Payment processing failed"
		p.AddLog("Payment Failed", oldStatus.String(), p.Status.String(), p.FailureReason)
	}

	return nil
}

// CanRefund 检查支付是否可以进行退款。
func (p *Payment) CanRefund() error {
	// 只有支付成功的订单才能退款。
	if p.Status != PaymentSuccess {
		return fmt.Errorf("payment status must be Success, current: %s", p.Status.String())
	}
	return nil
}

// CreateRefund 创建一笔退款请求。
// refundAmount: 退款金额。
// reason: 退款原因。
func (p *Payment) CreateRefund(refundAmount int64, reason string) (*Refund, error) {
	if err := p.CanRefund(); err != nil {
		return nil, err
	}

	if refundAmount <= 0 || refundAmount > p.Amount {
		return nil, fmt.Errorf("invalid refund amount")
	}

	refund := &Refund{
		RefundNo:     generateRefundNo(), // 生成唯一的退款单号。
		PaymentID:    p.ID,
		PaymentNo:    p.PaymentNo,
		OrderID:      p.OrderID,
		OrderNo:      p.OrderNo,
		UserID:       p.UserID,
		RefundAmount: refundAmount,
		Reason:       reason,
		Status:       PaymentRefunding, // 初始状态为退款中。
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	p.Refunds = append(p.Refunds, refund) // 将退款添加到支付的退款列表中。
	p.Status = PaymentRefunding           // 支付状态变更为退款中。
	p.UpdatedAt = time.Now()

	p.AddLog("Refund Created", "", "", fmt.Sprintf("Refund amount: %d, Reason: %s", refundAmount, reason))

	return refund, nil
}

// ProcessRefund 处理退款回调的结果。
// refundNo: 退款单号。
// success: 退款是否成功。
func (p *Payment) ProcessRefund(refundNo string, success bool) error {
	var refund *Refund
	// 查找对应的退款记录。
	for _, r := range p.Refunds {
		if r.RefundNo == refundNo {
			refund = r
			break
		}
	}

	if refund == nil {
		return fmt.Errorf("refund not found for payment %s", p.PaymentNo)
	}

	refund.UpdatedAt = time.Now()

	if success {
		refund.Status = PaymentRefunded // 退款成功。
		now := time.Now()
		refund.RefundedAt = &now // 记录退款成功时间。
		// 如果是全额退款，可以将支付状态也标记为已退款。
		// 这里简化处理，直接标记支付状态为已退款。
		p.Status = PaymentRefunded
		p.RefundedAt = &now
		p.AddLog("Refund Success", "", PaymentRefunded.String(), fmt.Sprintf("RefundNo: %s", refundNo))
	} else {
		refund.Status = PaymentFailed // 退款失败。
		refund.FailureReason = "Refund processing failed"
		p.AddLog("Refund Failed", "", PaymentFailed.String(), fmt.Sprintf("RefundNo: %s, Reason: %s", refundNo, refund.FailureReason))
	}

	p.UpdatedAt = time.Now()

	return nil
}

// Cancel 取消支付。
// reason: 取消原因。
func (p *Payment) Cancel(reason string) error {
	// 只有待处理状态的支付才能取消。
	if p.Status != PaymentPending {
		return fmt.Errorf("payment cannot be cancelled in current status: %s", p.Status.String())
	}

	oldStatus := p.Status
	p.Status = PaymentCancelled // 状态变更为已取消。
	now := time.Now()
	p.CancelledAt = &now // 记录取消时间。
	p.UpdatedAt = now
	p.FailureReason = reason

	p.AddLog("Payment Cancelled", oldStatus.String(), p.Status.String(), reason)

	return nil
}

// AddLog 添加支付操作日志。
// action: 操作动作。
// oldStatus: 旧状态。
// newStatus: 新状态。
// remark: 备注信息。
func (p *Payment) AddLog(action, oldStatus, newStatus, remark string) {
	log := &PaymentLog{
		PaymentID: p.ID,
		Action:    action,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Remark:    remark,
		CreatedAt: time.Now(),
	}
	p.Logs = append(p.Logs, log) // 将新日志添加到日志列表中。
}

// generatePaymentNo 生成唯一的支付单号。
// 格式为 PYYYYMMDDHHMMSS。
func generatePaymentNo() string {
	return fmt.Sprintf("P%s", time.Now().Format("20060102150405"))
}

// generateRefundNo 生成唯一的退款单号。
// 格式为 RYYYYMMDDHHMMSS。
func generateRefundNo() string {
	return fmt.Sprintf("R%s", time.Now().Format("20060102150405"))
}

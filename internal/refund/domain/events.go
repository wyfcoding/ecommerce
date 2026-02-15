// Package domain 退款领域事件
package domain

import "time"

// RefundCreatedEvent 退款申请创建事件
type RefundCreatedEvent struct {
	RefundID   uint64    `json:"refund_id"`
	RefundNo   string    `json:"refund_no"`
	OrderID    string    `json:"order_id"`
	Amount     int64     `json:"amount"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
}

func (e *RefundCreatedEvent) EventName() string { return "refund.created" }
func (e *RefundCreatedEvent) EventKey() string  { return e.RefundNo }

// RefundApprovedEvent 退款审核通过事件（准备打款）
type RefundApprovedEvent struct {
	RefundID   uint64    `json:"refund_id"`
	RefundNo   string    `json:"refund_no"`
	PaymentID  string    `json:"payment_id"`
	Amount     int64     `json:"amount"`
	ApprovedAt time.Time `json:"approved_at"`
}

func (e *RefundApprovedEvent) EventName() string { return "refund.approved" }
func (e *RefundApprovedEvent) EventKey() string  { return e.RefundNo }

// RefundSucceededEvent 退款成功事件
type RefundSucceededEvent struct {
	RefundID   uint64    `json:"refund_id"`
	RefundNo   string    `json:"refund_no"`
	OrderID    string    `json:"order_id"`
	Amount     int64     `json:"amount"`
	SucceededAt time.Time `json:"succeeded_at"`
}

func (e *RefundSucceededEvent) EventName() string { return "refund.succeeded" }
func (e *RefundSucceededEvent) EventKey() string  { return e.RefundNo }

// RefundFailedEvent 退款失败事件
type RefundFailedEvent struct {
	RefundID   uint64    `json:"refund_id"`
	RefundNo   string    `json:"refund_no"`
	Reason     string    `json:"reason"`
	FailedAt   time.Time `json:"failed_at"`
}

func (e *RefundFailedEvent) EventName() string { return "refund.failed" }
func (e *RefundFailedEvent) EventKey() string  { return e.RefundNo }

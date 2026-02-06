package domain

import (
	"time"

	"github.com/wyfcoding/pkg/eventsourcing"
)

// 定义事件类型常量。
const (
	PaymentInitiatedEventType       = "payment.initiated"
	PaymentAuthorizedEventType      = "payment.authorized"
	PaymentCapturedEventType        = "payment.captured"
	PaymentPaidEventType            = "payment.paid"
	PaymentClosedEventType          = "payment.closed"
	PaymentRefundedEventType        = "payment.refunded"
	PaymentReconciledEventType      = "payment.reconciled"
	PaymentReconcileFailedEventType = "payment.reconcile_failed"
	RefundInitiatedEventType        = "refund.initiated"
	RefundFinishedEventType         = "refund.finished"
)

// PaymentInitiatedEvent 当支付单初始化时发出的事件。
type PaymentInitiatedEvent struct {
	eventsourcing.BaseEvent
	OrderNo       string `json:"order_no"`
	OrderID       uint64 `json:"order_id"`
	UserID        uint64 `json:"user_id"`
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
}

func (e *PaymentInitiatedEvent) EventType() string { return PaymentInitiatedEventType }

// PaymentAuthorizedEvent 当网关授权成功时发出的事件。
type PaymentAuthorizedEvent struct {
	eventsourcing.BaseEvent
	PaymentNo     string `json:"payment_no"`
	OrderID       uint64 `json:"order_id"`
	UserID        uint64 `json:"user_id"`
	TransactionID string `json:"transaction_id"`
}

func (e *PaymentAuthorizedEvent) EventType() string { return PaymentAuthorizedEventType }

// PaymentCapturedEvent 当支付成功捕获时发出的事件。
type PaymentCapturedEvent struct {
	eventsourcing.BaseEvent
	PaymentNo string    `json:"payment_no"`
	OrderNo   string    `json:"order_no"`
	OrderID   uint64    `json:"order_id"`
	UserID    uint64    `json:"user_id"`
	Amount    int64     `json:"amount"`
	PaidAt    time.Time `json:"paid_at"`
}

func (e *PaymentCapturedEvent) EventType() string { return PaymentCapturedEventType }

// PaymentPaidEvent 当支付完成回调时发出的事件。
type PaymentPaidEvent struct {
	eventsourcing.BaseEvent
	PaymentNo string `json:"payment_no"`
	OrderNo   string `json:"order_no"`
	OrderID   uint64 `json:"order_id"`
	UserID    uint64 `json:"user_id"`
	Amount    int64  `json:"amount"`
	PaidAt    int64  `json:"paid_at"`
}

func (e *PaymentPaidEvent) EventType() string { return PaymentPaidEventType }

// RefundFinishedEvent 当退款完成时发出的事件。
type RefundFinishedEvent struct {
	eventsourcing.BaseEvent
	RefundNo     string    `json:"refund_no"`
	PaymentNo    string    `json:"payment_no"`
	OrderNo      string    `json:"order_no"`
	OrderID      uint64    `json:"order_id"`
	UserID       uint64    `json:"user_id"`
	RefundAmount int64     `json:"refund_amount"`
	RefundedAt   time.Time `json:"refunded_at"`
}

func (e *RefundFinishedEvent) EventType() string { return RefundFinishedEventType }

// PaymentClosedEvent 当支付单关闭时发出的事件。
type PaymentClosedEvent struct {
	eventsourcing.BaseEvent
	PaymentNo string    `json:"payment_no"`
	OrderID   uint64    `json:"order_id"`
	UserID    uint64    `json:"user_id"`
	ClosedAt  time.Time `json:"closed_at"`
}

func (e *PaymentClosedEvent) EventType() string { return PaymentClosedEventType }

// PaymentRefundedEvent 当支付完全退款时发出的事件。
type PaymentRefundedEvent struct {
	eventsourcing.BaseEvent
	RefundedAt time.Time `json:"refunded_at"`
}

func (e *PaymentRefundedEvent) EventType() string { return PaymentRefundedEventType }

// PaymentReconciledEvent 当支付对账成功时发出的事件。
type PaymentReconciledEvent struct {
	eventsourcing.BaseEvent
}

func (e *PaymentReconciledEvent) EventType() string { return PaymentReconciledEventType }

// PaymentReconcileFailedEvent 当支付对账失败时发出的事件。
type PaymentReconcileFailedEvent struct {
	eventsourcing.BaseEvent
}

func (e *PaymentReconcileFailedEvent) EventType() string { return PaymentReconcileFailedEventType }

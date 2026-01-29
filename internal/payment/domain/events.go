package domain

import "time"

// PaymentCapturedEvent 当支付成功捕获时发出的事件。
type PaymentCapturedEvent struct {
	PaymentNo string    `json:"payment_no"`
	OrderNo   string    `json:"order_no"`
	UserID    uint64    `json:"user_id"`
	Amount    int64     `json:"amount"`
	Timestamp int64     `json:"timestamp"`
	PaidAt    time.Time `json:"paid_at"`
}

// PaymentPaidEvent 当支付完成回调时发出的事件。
type PaymentPaidEvent struct {
	PaymentNo string `json:"payment_no"`
	OrderNo   string `json:"order_no"`
	UserID    uint64 `json:"user_id"`
	Amount    int64  `json:"amount"`
	PaidAt    int64  `json:"paid_at"`
}

// RefundFinishedEvent 当退款完成时发出的事件。
type RefundFinishedEvent struct {
	RefundNo     string    `json:"refund_no"`
	PaymentNo    string    `json:"payment_no"`
	OrderNo      string    `json:"order_no"`
	UserID       uint64    `json:"user_id"`
	RefundAmount int64     `json:"refund_amount"`
	RefundedAt   time.Time `json:"refunded_at"`
}

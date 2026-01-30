package domain

import "time"

// AfterSalesCreatedEvent 售后申请创建事件。
type AfterSalesCreatedEvent struct {
	AfterSalesID uint64    `json:"after_sales_id"`
	OrderID      uint64    `json:"order_id"`
	UserID       uint64    `json:"user_id"`
	Type         string    `json:"type"` // REFUND, RETURN, REPLACEMENT
	Reason       string    `json:"reason"`
	Timestamp    time.Time `json:"timestamp"`
}

// AfterSalesStatusUpdatedEvent 售后状态更新事件。
type AfterSalesStatusUpdatedEvent struct {
	AfterSalesID uint64    `json:"after_sales_id"`
	OldStatus    string    `json:"old_status"`
	NewStatus    string    `json:"new_status"`
	OperatorID   uint64    `json:"operator_id"`
	Timestamp    time.Time `json:"timestamp"`
}

// RefundProcessedEvent 退款处理完成事件。
type RefundProcessedEvent struct {
	AfterSalesID uint64    `json:"after_sales_id"`
	RefundID     string    `json:"refund_id"`
	Amount       int64     `json:"amount"`
	Channel      string    `json:"channel"`
	Timestamp    time.Time `json:"timestamp"`
}

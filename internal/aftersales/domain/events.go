// 变更说明：
// 1. 新增退款相关事件（RefundInitiated/RefundSuccess/RefundFailed）
// 2. 新增退货物流相关事件（ReturnShipped/ReturnReceived/QCCompleted）
package domain

import "time"

const (
	// AfterSalesCreatedEventType 售后申请创建事件。
	AfterSalesCreatedEventType = "aftersales.created"
	// AfterSalesStatusUpdatedEventType 售后状态更新事件。
	AfterSalesStatusUpdatedEventType = "aftersales.status.updated"
	// AfterSalesSupportTicketCreatedType 客服工单创建事件。
	AfterSalesSupportTicketCreatedType = "aftersales.support.ticket.created"
	// AfterSalesSupportTicketUpdatedType 客服工单更新事件。
	AfterSalesSupportTicketUpdatedType = "aftersales.support.ticket.updated"
	// AfterSalesSupportTicketMessageType 客服工单消息创建事件。
	AfterSalesSupportTicketMessageType = "aftersales.support.ticket.message.created"
	// AfterSalesConfigUpdatedEventType 售后配置更新事件。
	AfterSalesConfigUpdatedEventType = "aftersales.config.updated"
	// AfterSalesRefundInitiatedEventType 退款发起事件。
	AfterSalesRefundInitiatedEventType = "aftersales.refund.initiated"
	// AfterSalesRefundSuccessEventType 退款成功事件。
	AfterSalesRefundSuccessEventType = "aftersales.refund.success"
	// AfterSalesRefundFailedEventType 退款失败事件。
	AfterSalesRefundFailedEventType = "aftersales.refund.failed"
	// AfterSalesReturnShippedEventType 退货寄出事件。
	AfterSalesReturnShippedEventType = "aftersales.return.shipped"
	// AfterSalesReturnReceivedEventType 退货签收事件。
	AfterSalesReturnReceivedEventType = "aftersales.return.received"
	// AfterSalesQCCompletedEventType 质检完成事件。
	AfterSalesQCCompletedEventType = "aftersales.qc.completed"
)

// AfterSalesCreatedEvent 售后申请创建事件。
type AfterSalesCreatedEvent struct {
	AfterSalesID uint64           `json:"after_sales_id"`
	AfterSalesNo string           `json:"after_sales_no"`
	OrderID      uint64           `json:"order_id"`
	UserID       uint64           `json:"user_id"`
	Type         AfterSalesType   `json:"type"`
	Status       AfterSalesStatus `json:"status"`
	Timestamp    time.Time        `json:"timestamp"`
}

// AfterSalesStatusUpdatedEvent 售后状态更新事件。
type AfterSalesStatusUpdatedEvent struct {
	AfterSalesID uint64           `json:"after_sales_id"`
	AfterSalesNo string           `json:"after_sales_no"`
	OldStatus    AfterSalesStatus `json:"old_status"`
	NewStatus    AfterSalesStatus `json:"new_status"`
	Operator     string           `json:"operator"`
	Timestamp    time.Time        `json:"timestamp"`
}

// SupportTicketCreatedEvent 客服工单创建事件。
type SupportTicketCreatedEvent struct {
	TicketID  uint64    `json:"ticket_id"`
	TicketNo  string    `json:"ticket_no"`
	UserID    uint64    `json:"user_id"`
	OrderID   uint64    `json:"order_id"`
	Timestamp time.Time `json:"timestamp"`
}

// SupportTicketUpdatedEvent 客服工单更新事件。
type SupportTicketUpdatedEvent struct {
	TicketID  uint64              `json:"ticket_id"`
	Status    SupportTicketStatus `json:"status"`
	Timestamp time.Time           `json:"timestamp"`
}

// SupportTicketMessageCreatedEvent 客服工单消息创建事件。
type SupportTicketMessageCreatedEvent struct {
	MessageID uint64    `json:"message_id"`
	TicketID  uint64    `json:"ticket_id"`
	SenderID  uint64    `json:"sender_id"`
	Timestamp time.Time `json:"timestamp"`
}

// AfterSalesConfigUpdatedEvent 售后配置更新事件。
type AfterSalesConfigUpdatedEvent struct {
	Key       string    `json:"key"`
	Timestamp time.Time `json:"timestamp"`
}

// RefundInitiatedEvent 退款发起事件。
// 下游消费者：支付服务（执行退款）、通知服务（发送退款通知）。
type RefundInitiatedEvent struct {
	AfterSalesID uint64        `json:"after_sales_id"`
	AfterSalesNo string        `json:"after_sales_no"`
	RefundNo     string        `json:"refund_no"`
	PaymentID    string        `json:"payment_id"`
	Amount       int64         `json:"amount"`
	Currency     string        `json:"currency"`
	Channel      RefundChannel `json:"channel"`
	UserID       uint64        `json:"user_id"`
	Timestamp    time.Time     `json:"timestamp"`
}

// RefundSuccessEvent 退款成功事件。
// 下游消费者：库存服务（释放库存）、通知服务（发送退款成功通知）。
type RefundSuccessEvent struct {
	AfterSalesID uint64    `json:"after_sales_id"`
	RefundNo     string    `json:"refund_no"`
	ChannelTxID  string    `json:"channel_tx_id"`
	Amount       int64     `json:"amount"`
	UserID       uint64    `json:"user_id"`
	Timestamp    time.Time `json:"timestamp"`
}

// RefundFailedEvent 退款失败事件。
// 下游消费者：告警服务（触发人工介入告警）。
type RefundFailedEvent struct {
	AfterSalesID uint64    `json:"after_sales_id"`
	RefundNo     string    `json:"refund_no"`
	ErrorMessage string    `json:"error_message"`
	RetryCount   int       `json:"retry_count"`
	MaxRetry     int       `json:"max_retry"`
	Timestamp    time.Time `json:"timestamp"`
}

// ReturnShippedEvent 退货寄出事件。
// 下游消费者：物流服务（追踪物流状态）。
type ReturnShippedEvent struct {
	AfterSalesID     uint64    `json:"after_sales_id"`
	TrackingNumber   string    `json:"tracking_number"`
	LogisticsCompany string    `json:"logistics_company"`
	Deadline         time.Time `json:"deadline"`
	Timestamp        time.Time `json:"timestamp"`
}

// ReturnReceivedEvent 退货签收事件。
// 下游消费者：仓库服务（触发质检流程）。
type ReturnReceivedEvent struct {
	AfterSalesID uint64    `json:"after_sales_id"`
	ReceivedAt   time.Time `json:"received_at"`
	Timestamp    time.Time `json:"timestamp"`
}

// QCCompletedEvent 质检完成事件。
// 下游消费者：售后服务（根据质检结果决定退款或拒绝）。
type QCCompletedEvent struct {
	AfterSalesID uint64    `json:"after_sales_id"`
	Passed       bool      `json:"passed"`
	QCResult     string    `json:"qc_result"`
	Notes        string    `json:"notes"`
	Timestamp    time.Time `json:"timestamp"`
}

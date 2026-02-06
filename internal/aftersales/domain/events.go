package domain

import "time"

const (
	AfterSalesCreatedEventType         = "aftersales.created"
	AfterSalesStatusUpdatedEventType   = "aftersales.status.updated"
	AfterSalesSupportTicketCreatedType = "aftersales.support.ticket.created"
	AfterSalesSupportTicketUpdatedType = "aftersales.support.ticket.updated"
	AfterSalesSupportTicketMessageType = "aftersales.support.ticket.message.created"
	AfterSalesConfigUpdatedEventType   = "aftersales.config.updated"
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

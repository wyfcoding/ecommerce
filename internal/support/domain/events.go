package domain

import "time"

const (
	TicketCreatedEventType              = "support.ticket.created"
	TicketUpdatedEventType              = "support.ticket.updated"
	TicketMessageCreatedEventType       = "support.ticket.message.created"
	ConversationCreatedEventType        = "support.conversation.created"
	ConversationMessageCreatedEventType = "support.conversation.message.created"
)

// TicketCreatedEvent 工单创建事件。
type TicketCreatedEvent struct {
	TicketID  uint64    `json:"ticket_id"`
	TicketNo  string    `json:"ticket_no"`
	UserID    uint64    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

// TicketUpdatedEvent 工单状态/分配更新事件。
type TicketUpdatedEvent struct {
	TicketID   uint64       `json:"ticket_id"`
	Status     TicketStatus `json:"status"`
	AssigneeID uint64       `json:"assignee_id"`
	Timestamp  time.Time    `json:"timestamp"`
}

// TicketMessageCreatedEvent 工单消息创建事件。
type TicketMessageCreatedEvent struct {
	MessageID uint64    `json:"message_id"`
	TicketID  uint64    `json:"ticket_id"`
	SenderID  uint64    `json:"sender_id"`
	Timestamp time.Time `json:"timestamp"`
}

// ConversationCreatedEvent 会话创建事件。
type ConversationCreatedEvent struct {
	ConversationID uint64    `json:"conversation_id"`
	User1ID        uint64    `json:"user1_id"`
	User2ID        uint64    `json:"user2_id"`
	Timestamp      time.Time `json:"timestamp"`
}

// ConversationMessageCreatedEvent 会话消息创建事件。
type ConversationMessageCreatedEvent struct {
	MessageID      uint64    `json:"message_id"`
	ConversationID uint64    `json:"conversation_id"`
	SenderID       uint64    `json:"sender_id"`
	ReceiverID     uint64    `json:"receiver_id"`
	Timestamp      time.Time `json:"timestamp"`
}

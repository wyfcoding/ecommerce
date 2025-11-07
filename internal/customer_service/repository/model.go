package repository

import (
	"gorm.io/gorm"
)

// Ticket represents a customer service ticket.
type Ticket struct {
	gorm.Model
	TicketID    string `gorm:"uniqueIndex;not null;comment:工单唯一ID" json:"ticketId"`
	UserID      uint64 `gorm:"index;not null;comment:用户ID" json:"userId"`
	Subject     string `gorm:"size:255;not null;comment:工单主题" json:"subject"`
	Description string `gorm:"type:text;comment:工单描述" json:"description"`
	Status      string `gorm:"size:50;not null;comment:工单状态 (OPEN, IN_PROGRESS, RESOLVED, CLOSED)" json:"status"`
	// Add other fields like assigned_agent_id, priority, etc.
}

// TicketMessage represents a message within a customer service ticket.
type TicketMessage struct {
	gorm.Model
	MessageID  string `gorm:"uniqueIndex;not null;comment:消息唯一ID" json:"messageId"`
	TicketID   string `gorm:"index;not null;comment:关联工单ID" json:"ticketId"`
	SenderID   uint64 `gorm:"not null;comment:发送者ID (用户或客服)" json:"senderId"`
	SenderType string `gorm:"size:20;not null;comment:发送者类型 (USER, AGENT)" json:"senderType"`
	Content    string `gorm:"type:text;not null;comment:消息内容" json:"content"`
	// Add other fields like attachments, read_status, etc.
}

// TableName specifies the table name for Ticket.
func (Ticket) TableName() string {
	return "tickets"
}

// TableName specifies the table name for TicketMessage.
func (TicketMessage) TableName() string {
	return "ticket_messages"
}

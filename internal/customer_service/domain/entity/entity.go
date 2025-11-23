package entity

import (
	"time"

	"gorm.io/gorm"
)

// TicketStatus 工单状态
type TicketStatus int

const (
	TicketStatusOpen       TicketStatus = 1 // 待处理
	TicketStatusInProgress TicketStatus = 2 // 处理中
	TicketStatusResolved   TicketStatus = 3 // 已解决
	TicketStatusClosed     TicketStatus = 4 // 已关闭
)

// TicketPriority 工单优先级
type TicketPriority int

const (
	TicketPriorityLow    TicketPriority = 1 // 低
	TicketPriorityMedium TicketPriority = 2 // 中
	TicketPriorityHigh   TicketPriority = 3 // 高
	TicketPriorityUrgent TicketPriority = 4 // 紧急
)

// Ticket 工单实体
type Ticket struct {
	gorm.Model
	TicketNo    string         `gorm:"type:varchar(64);uniqueIndex;not null;comment:工单编号" json:"ticket_no"`
	UserID      uint64         `gorm:"not null;index;comment:用户ID" json:"user_id"`
	Subject     string         `gorm:"type:varchar(255);not null;comment:主题" json:"subject"`
	Description string         `gorm:"type:text;comment:描述" json:"description"`
	Status      TicketStatus   `gorm:"default:1;comment:状态" json:"status"`
	Priority    TicketPriority `gorm:"default:2;comment:优先级" json:"priority"`
	Category    string         `gorm:"type:varchar(64);comment:分类" json:"category"`
	AssigneeID  uint64         `gorm:"comment:经办人ID" json:"assignee_id"`
	ResolvedAt  *time.Time     `gorm:"comment:解决时间" json:"resolved_at"`
	ClosedAt    *time.Time     `gorm:"comment:关闭时间" json:"closed_at"`
}

// MessageType 消息类型
type MessageType int

const (
	MessageTypeText  MessageType = 1 // 文本
	MessageTypeImage MessageType = 2 // 图片
	MessageTypeFile  MessageType = 3 // 文件
)

// Message 工单消息实体
type Message struct {
	gorm.Model
	TicketID   uint64      `gorm:"not null;index;comment:工单ID" json:"ticket_id"`
	SenderID   uint64      `gorm:"not null;comment:发送者ID" json:"sender_id"`
	SenderType string      `gorm:"type:varchar(32);not null;comment:发送者类型(user/admin)" json:"sender_type"`
	Content    string      `gorm:"type:text;not null;comment:内容" json:"content"`
	Type       MessageType `gorm:"default:1;comment:消息类型" json:"type"`
	IsInternal bool        `gorm:"default:false;comment:是否内部消息" json:"is_internal"`
}

// NewTicket 创建工单
func NewTicket(ticketNo string, userID uint64, subject, description, category string, priority TicketPriority) *Ticket {
	return &Ticket{
		TicketNo:    ticketNo,
		UserID:      userID,
		Subject:     subject,
		Description: description,
		Category:    category,
		Priority:    priority,
		Status:      TicketStatusOpen,
	}
}

// NewMessage 创建消息
func NewMessage(ticketID, senderID uint64, senderType, content string, msgType MessageType, isInternal bool) *Message {
	return &Message{
		TicketID:   ticketID,
		SenderID:   senderID,
		SenderType: senderType,
		Content:    content,
		Type:       msgType,
		IsInternal: isInternal,
	}
}

// Assign 分配工单
func (t *Ticket) Assign(assigneeID uint64) {
	t.AssigneeID = assigneeID
	if t.Status == TicketStatusOpen {
		t.Status = TicketStatusInProgress
	}
}

// Resolve 解决工单
func (t *Ticket) Resolve() {
	t.Status = TicketStatusResolved
	now := time.Now()
	t.ResolvedAt = &now
}

// Close 关闭工单
func (t *Ticket) Close() {
	t.Status = TicketStatusClosed
	now := time.Now()
	t.ClosedAt = &now
}

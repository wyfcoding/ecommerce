package domain

import "time"

// TicketStatus 定义了工单的生命周期状态。
type TicketStatus int

const (
	TicketStatusOpen       TicketStatus = 1 // 待处理。
	TicketStatusInProgress TicketStatus = 2 // 处理中。
	TicketStatusResolved   TicketStatus = 3 // 已解决。
	TicketStatusClosed     TicketStatus = 4 // 已关闭。
)

// TicketPriority 定义了工单的优先级。
type TicketPriority int

const (
	TicketPriorityLow    TicketPriority = 1 // 低。
	TicketPriorityMedium TicketPriority = 2 // 中。
	TicketPriorityHigh   TicketPriority = 3 // 高。
	TicketPriorityUrgent TicketPriority = 4 // 紧急。
)

// Ticket 实体代表一个客户服务工单。
type Ticket struct {
	ID          uint64         `json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	TicketNo    string         `json:"ticket_no"`
	UserID      uint64         `json:"user_id"`
	Subject     string         `json:"subject"`
	Description string         `json:"description"`
	Status      TicketStatus   `json:"status"`
	Priority    TicketPriority `json:"priority"`
	Category    string         `json:"category"`
	AssigneeID  uint64         `json:"assignee_id"`
	ResolvedAt  *time.Time     `json:"resolved_at"`
	ClosedAt    *time.Time     `json:"closed_at"`
}

// MessageType 定义了工单消息的类型。
type MessageType int

const (
	MessageTypeText  MessageType = 1 // 文本消息。
	MessageTypeImage MessageType = 2 // 图片消息。
	MessageTypeFile  MessageType = 3 // 文件消息。
)

// Message 实体代表工单中的一条消息。
type Message struct {
	ID         uint64      `json:"id"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	TicketID   uint64      `json:"ticket_id"`
	SenderID   uint64      `json:"sender_id"`
	SenderType string      `json:"sender_type"`
	Content    string      `json:"content"`
	Type       MessageType `json:"type"`
	IsInternal bool        `json:"is_internal"`
}

// NewTicket 创建并返回一个新的 Ticket 实体实例。
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

// NewMessage 创建并返回一个新的 Message 实体实例。
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

// Assign 为工单分配经办人。
func (t *Ticket) Assign(assigneeID uint64) {
	t.AssigneeID = assigneeID
	if t.Status == TicketStatusOpen {
		t.Status = TicketStatusInProgress
	}
}

// Resolve 解决工单。
func (t *Ticket) Resolve() {
	t.Status = TicketStatusResolved
	now := time.Now()
	t.ResolvedAt = &now
}

// Close 关闭工单。
func (t *Ticket) Close() {
	t.Status = TicketStatusClosed
	now := time.Now()
	t.ClosedAt = &now
}

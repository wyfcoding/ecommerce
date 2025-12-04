package entity

import (
	"time"

	"gorm.io/gorm" // 导入GORM库。
)

// TicketStatus 定义了工单的生命周期状态。
type TicketStatus int

const (
	TicketStatusOpen       TicketStatus = 1 // 待处理：工单已创建，等待分配或处理。
	TicketStatusInProgress TicketStatus = 2 // 处理中：工单已分配给经办人，正在处理中。
	TicketStatusResolved   TicketStatus = 3 // 已解决：问题已解决，等待客户确认或自动关闭。
	TicketStatusClosed     TicketStatus = 4 // 已关闭：工单已完成并关闭。
)

// TicketPriority 定义了工单的优先级。
type TicketPriority int

const (
	TicketPriorityLow    TicketPriority = 1 // 低：优先级较低的工单。
	TicketPriorityMedium TicketPriority = 2 // 中：中等优先级的工单。
	TicketPriorityHigh   TicketPriority = 3 // 高：高优先级的工单。
	TicketPriorityUrgent TicketPriority = 4 // 紧急：需要立即处理的紧急工单。
)

// Ticket 实体代表一个客户服务工单。
// 它包含了工单的基本信息、状态、优先级、分配情况以及解决和关闭时间。
type Ticket struct {
	gorm.Model                 // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	TicketNo    string         `gorm:"type:varchar(64);uniqueIndex;not null;comment:工单编号" json:"ticket_no"` // 工单的唯一编号，唯一索引，不允许为空。
	UserID      uint64         `gorm:"not null;index;comment:用户ID" json:"user_id"`                          // 提交工单的用户ID，索引字段。
	Subject     string         `gorm:"type:varchar(255);not null;comment:主题" json:"subject"`                // 工单主题。
	Description string         `gorm:"type:text;comment:描述" json:"description"`                             // 工单详细描述。
	Status      TicketStatus   `gorm:"default:1;comment:状态" json:"status"`                                  // 工单状态，默认为待处理。
	Priority    TicketPriority `gorm:"default:2;comment:优先级" json:"priority"`                               // 工单优先级，默认为中。
	Category    string         `gorm:"type:varchar(64);comment:分类" json:"category"`                         // 工单分类，例如“订单问题”、“技术支持”。
	AssigneeID  uint64         `gorm:"comment:经办人ID" json:"assignee_id"`                                    // 分配给处理该工单的经办人（客服或管理员）ID。
	ResolvedAt  *time.Time     `gorm:"comment:解决时间" json:"resolved_at"`                                     // 工单解决时间。
	ClosedAt    *time.Time     `gorm:"comment:关闭时间" json:"closed_at"`                                       // 工单关闭时间。
}

// MessageType 定义了工单消息的类型。
type MessageType int

const (
	MessageTypeText  MessageType = 1 // 文本消息。
	MessageTypeImage MessageType = 2 // 图片消息。
	MessageTypeFile  MessageType = 3 // 文件消息。
)

// Message 实体代表工单中的一条消息。
// 它记录了发送者、消息内容、类型以及是否为内部消息等信息。
type Message struct {
	gorm.Model             // 嵌入gorm.Model。
	TicketID   uint64      `gorm:"not null;index;comment:工单ID" json:"ticket_id"`                           // 关联的工单ID，索引字段。
	SenderID   uint64      `gorm:"not null;comment:发送者ID" json:"sender_id"`                                // 消息发送者的ID。
	SenderType string      `gorm:"type:varchar(32);not null;comment:发送者类型(user/admin)" json:"sender_type"` // 发送者类型，例如“user”或“admin”。
	Content    string      `gorm:"type:text;not null;comment:内容" json:"content"`                           // 消息内容。
	Type       MessageType `gorm:"default:1;comment:消息类型" json:"type"`                                     // 消息类型，默认为文本。
	IsInternal bool        `gorm:"default:false;comment:是否内部消息" json:"is_internal"`                        // 标记是否为内部消息（仅供客服团队内部查看）。
}

// NewTicket 创建并返回一个新的 Ticket 实体实例。
// ticketNo: 工单编号。
// userID: 提交用户ID。
// subject, description, category: 工单的主题、描述和分类。
// priority: 工单优先级。
func NewTicket(ticketNo string, userID uint64, subject, description, category string, priority TicketPriority) *Ticket {
	return &Ticket{
		TicketNo:    ticketNo,
		UserID:      userID,
		Subject:     subject,
		Description: description,
		Category:    category,
		Priority:    priority,
		Status:      TicketStatusOpen, // 新创建的工单默认为待处理状态。
	}
}

// NewMessage 创建并返回一个新的 Message 实体实例。
// ticketID: 关联的工单ID。
// senderID: 发送者ID。
// senderType: 发送者类型。
// content: 消息内容。
// msgType: 消息类型。
// isInternal: 是否为内部消息。
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

// Assign 为工单分配经办人，并根据情况更新工单状态。
// assigneeID: 经办人ID。
func (t *Ticket) Assign(assigneeID uint64) {
	t.AssigneeID = assigneeID
	// 如果工单当前状态是Open，则将其更新为InProgress。
	if t.Status == TicketStatusOpen {
		t.Status = TicketStatusInProgress
	}
}

// Resolve 解决工单，更新工单状态为“已解决”，并记录解决时间。
func (t *Ticket) Resolve() {
	t.Status = TicketStatusResolved // 状态更新为“已解决”。
	now := time.Now()
	t.ResolvedAt = &now // 记录解决时间。
}

// Close 关闭工单，更新工单状态为“已关闭”，并记录关闭时间。
func (t *Ticket) Close() {
	t.Status = TicketStatusClosed // 状态更新为“已关闭”。
	now := time.Now()
	t.ClosedAt = &now // 记录关闭时间。
}

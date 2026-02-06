package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/support/domain"
	"gorm.io/gorm"
)

// TicketModel 工单写模型。
type TicketModel struct {
	gorm.Model
	TicketNo    string                `gorm:"type:varchar(64);uniqueIndex;not null;comment:工单编号"`
	UserID      uint64                `gorm:"not null;index;comment:用户ID"`
	Subject     string                `gorm:"type:varchar(255);not null;comment:主题"`
	Description string                `gorm:"type:text;comment:描述"`
	Status      domain.TicketStatus   `gorm:"default:1;comment:状态"`
	Priority    domain.TicketPriority `gorm:"default:2;comment:优先级"`
	Category    string                `gorm:"type:varchar(64);comment:分类"`
	AssigneeID  uint64                `gorm:"comment:经办人ID"`
	ResolvedAt  *time.Time            `gorm:"comment:解决时间"`
	ClosedAt    *time.Time            `gorm:"comment:关闭时间"`
}

// TicketMessageModel 工单消息写模型。
type TicketMessageModel struct {
	gorm.Model
	TicketID   uint64             `gorm:"not null;index;comment:工单ID"`
	SenderID   uint64             `gorm:"not null;comment:发送者ID"`
	SenderType string             `gorm:"type:varchar(32);not null;comment:发送者类型(user/admin)"`
	Content    string             `gorm:"type:text;not null;comment:内容"`
	Type       domain.MessageType `gorm:"default:1;comment:消息类型"`
	IsInternal bool               `gorm:"default:false;comment:是否内部消息"`
}

// ConversationModel 会话写模型。
type ConversationModel struct {
	gorm.Model
	User1ID       uint64    `gorm:"index:idx_user1;not null;comment:用户1ID"`
	User2ID       uint64    `gorm:"index:idx_user2;not null;comment:用户2ID"`
	LastMessageID uint64    `gorm:"not null;comment:最后一条消息ID"`
	LastMessage   string    `gorm:"type:varchar(255);comment:最后一条消息内容"`
	LastMessageAt time.Time `gorm:"not null;comment:最后一条消息时间"`
	UnreadCount1  int32     `gorm:"not null;default:0;comment:用户1未读数"`
	UnreadCount2  int32     `gorm:"not null;default:0;comment:用户2未读数"`
}

// ConversationMessageModel 会话消息写模型。
type ConversationMessageModel struct {
	gorm.Model
	ConversationID uint64             `gorm:"index;not null;comment:会话ID"`
	SenderID       uint64             `gorm:"index;not null;comment:发送者ID"`
	ReceiverID     uint64             `gorm:"index;not null;comment:接收者ID"`
	Type           domain.MessageType `gorm:"default:1;comment:消息类型"`
	Content        string             `gorm:"type:text;not null;comment:内容"`
	IsRead         bool               `gorm:"default:false;comment:是否已读"`
	ReadAt         *time.Time         `gorm:"comment:阅读时间"`
}

func (TicketModel) TableName() string {
	return "tickets"
}

func (TicketMessageModel) TableName() string {
	return "messages"
}

func (ConversationModel) TableName() string {
	return "conversations"
}

func (ConversationMessageModel) TableName() string {
	return "conversation_messages"
}

func toTicketModel(t *domain.Ticket) *TicketModel {
	if t == nil {
		return nil
	}
	return &TicketModel{
		Model: gorm.Model{
			ID:        uint(t.ID),
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
		},
		TicketNo:    t.TicketNo,
		UserID:      t.UserID,
		Subject:     t.Subject,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		Category:    t.Category,
		AssigneeID:  t.AssigneeID,
		ResolvedAt:  t.ResolvedAt,
		ClosedAt:    t.ClosedAt,
	}
}

func toTicket(model *TicketModel) *domain.Ticket {
	if model == nil {
		return nil
	}
	return &domain.Ticket{
		ID:          uint64(model.ID),
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		TicketNo:    model.TicketNo,
		UserID:      model.UserID,
		Subject:     model.Subject,
		Description: model.Description,
		Status:      model.Status,
		Priority:    model.Priority,
		Category:    model.Category,
		AssigneeID:  model.AssigneeID,
		ResolvedAt:  model.ResolvedAt,
		ClosedAt:    model.ClosedAt,
	}
}

func toMessageModel(m *domain.Message) *TicketMessageModel {
	if m == nil {
		return nil
	}
	return &TicketMessageModel{
		Model: gorm.Model{
			ID:        uint(m.ID),
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		},
		TicketID:   m.TicketID,
		SenderID:   m.SenderID,
		SenderType: m.SenderType,
		Content:    m.Content,
		Type:       m.Type,
		IsInternal: m.IsInternal,
	}
}

func toMessage(model *TicketMessageModel) *domain.Message {
	if model == nil {
		return nil
	}
	return &domain.Message{
		ID:         uint64(model.ID),
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
		TicketID:   model.TicketID,
		SenderID:   model.SenderID,
		SenderType: model.SenderType,
		Content:    model.Content,
		Type:       model.Type,
		IsInternal: model.IsInternal,
	}
}

func toConversationModel(c *domain.Conversation) *ConversationModel {
	if c == nil {
		return nil
	}
	return &ConversationModel{
		Model: gorm.Model{
			ID:        uint(c.ID),
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		},
		User1ID:       c.User1ID,
		User2ID:       c.User2ID,
		LastMessageID: c.LastMessageID,
		LastMessage:   c.LastMessage,
		LastMessageAt: c.LastMessageAt,
		UnreadCount1:  c.UnreadCount1,
		UnreadCount2:  c.UnreadCount2,
	}
}

func toConversation(model *ConversationModel) *domain.Conversation {
	if model == nil {
		return nil
	}
	return &domain.Conversation{
		ID:            uint64(model.ID),
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
		User1ID:       model.User1ID,
		User2ID:       model.User2ID,
		LastMessageID: model.LastMessageID,
		LastMessage:   model.LastMessage,
		LastMessageAt: model.LastMessageAt,
		UnreadCount1:  model.UnreadCount1,
		UnreadCount2:  model.UnreadCount2,
	}
}

func toConversationMessageModel(m *domain.ConversationMessage) *ConversationMessageModel {
	if m == nil {
		return nil
	}
	return &ConversationMessageModel{
		Model: gorm.Model{
			ID:        uint(m.ID),
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		},
		ConversationID: m.ConversationID,
		SenderID:       m.SenderID,
		ReceiverID:     m.ReceiverID,
		Type:           m.Type,
		Content:        m.Content,
		IsRead:         m.IsRead,
		ReadAt:         m.ReadAt,
	}
}

func toConversationMessage(model *ConversationMessageModel) *domain.ConversationMessage {
	if model == nil {
		return nil
	}
	return &domain.ConversationMessage{
		ID:             uint64(model.ID),
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
		ConversationID: model.ConversationID,
		SenderID:       model.SenderID,
		ReceiverID:     model.ReceiverID,
		Type:           model.Type,
		Content:        model.Content,
		IsRead:         model.IsRead,
		ReadAt:         model.ReadAt,
	}
}

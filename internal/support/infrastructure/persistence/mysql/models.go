package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/support/domain"
	"gorm.io/gorm"
)

// --- Existing Ticket & P2P Models ---

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

// --- New Intelligent Support Models ---

// KnowledgeBaseModel 知识库写模型。
type KnowledgeBaseModel struct {
	gorm.Model
	KID         string `gorm:"column:k_id;type:varchar(64);uniqueIndex;not null;comment:知识库ID"`
	Name        string `gorm:"type:varchar(128);not null;comment:名称"`
	Description string `gorm:"type:varchar(255);comment:描述"`
	Language    string `gorm:"type:varchar(32);comment:语言"`
	IsActive    bool   `gorm:"default:true;comment:是否启用"`
}

// KnowledgeArticleModel 知识库文章写模型。
type KnowledgeArticleModel struct {
	gorm.Model
	AID             string `gorm:"column:a_id;type:varchar(64);uniqueIndex;not null;comment:文章ID"`
	KnowledgeBaseID string `gorm:"index;not null;comment:知识库ID"`
	Title           string `gorm:"type:varchar(255);not null;comment:标题"`
	Content         string `gorm:"type:text;not null;comment:内容"`
	Keywords        string `gorm:"type:varchar(255);comment:关键词(逗号分隔)"`
	Category        string `gorm:"type:varchar(64);comment:分类"`
	IsActive        bool   `gorm:"default:true;comment:是否启用"`
}

// AIConversationModel AI会话写模型。
type AIConversationModel struct {
	gorm.Model
	CID              string                    `gorm:"column:c_id;type:varchar(64);uniqueIndex;not null;comment:会话ID"`
	UserID           uint64                    `gorm:"index;not null;comment:用户ID"`
	Status           domain.ConversationStatus `gorm:"default:1;comment:状态"`
	PrimaryIntent    domain.IntentCategory     `gorm:"comment:主意图"`
	OverallSentiment domain.Sentiment          `gorm:"comment:整体情感"`
	StartedAt        time.Time                 `gorm:"not null;comment:开始时间"`
	EndedAt          *time.Time                `gorm:"comment:结束时间"`
}

// AIMessageModel AI消息写模型。
type AIMessageModel struct {
	gorm.Model
	MID            string                `gorm:"column:m_id;type:varchar(64);uniqueIndex;not null;comment:消息ID"`
	ConversationID string                `gorm:"index;not null;comment:会话ID"`
	Sender         domain.MessageSender  `gorm:"not null;comment:发送者"`
	Content        string                `gorm:"type:text;not null;comment:内容"`
	Sentiment      domain.Sentiment      `gorm:"comment:情感倾向"`
	Intent         domain.IntentCategory `gorm:"comment:意图"`
	Confidence     float64               `gorm:"comment:置信度"`
}

// --- Table Names ---

func (KnowledgeBaseModel) TableName() string {
	return "support_knowledge_bases"
}

func (KnowledgeArticleModel) TableName() string {
	return "support_knowledge_articles"
}

func (AIConversationModel) TableName() string {
	return "support_ai_conversations"
}

func (AIMessageModel) TableName() string {
	return "support_ai_messages"
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

// --- Conversion Functions ---

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

func toKnowledgeBaseModel(kb *domain.KnowledgeBase) *KnowledgeBaseModel {
	if kb == nil {
		return nil
	}
	return &KnowledgeBaseModel{
		Model: gorm.Model{
			UpdatedAt: kb.UpdatedAt,
			CreatedAt: kb.CreatedAt,
		},
		KID:         kb.ID,
		Name:        kb.Name,
		Description: kb.Description,
		Language:    kb.Language,
		IsActive:    kb.IsActive,
	}
}

func toKnowledgeBase(m *KnowledgeBaseModel) *domain.KnowledgeBase {
	if m == nil {
		return nil
	}
	return &domain.KnowledgeBase{
		ID:          m.KID,
		Name:        m.Name,
		Description: m.Description,
		Language:    m.Language,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func toAIConversationModel(c *domain.AIConversation) *AIConversationModel {
	if c == nil {
		return nil
	}
	return &AIConversationModel{
		Model: gorm.Model{
			CreatedAt: c.StartedAt,
		},
		CID:              c.ID,
		UserID:           c.UserID,
		Status:           c.Status,
		PrimaryIntent:    c.PrimaryIntent,
		OverallSentiment: c.OverallSentiment,
		StartedAt:        c.StartedAt,
		EndedAt:          c.EndedAt,
	}
}

func toAIConversation(m *AIConversationModel) *domain.AIConversation {
	if m == nil {
		return nil
	}
	return &domain.AIConversation{
		ID:               m.CID,
		UserID:           m.UserID,
		Status:           m.Status,
		PrimaryIntent:    m.PrimaryIntent,
		OverallSentiment: m.OverallSentiment,
		StartedAt:        m.StartedAt,
		EndedAt:          m.EndedAt,
	}
}

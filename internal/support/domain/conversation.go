package domain

import (
	"time"

	"gorm.io/gorm"
)

// MessageType definitions are reused from ticket.go

// Conversation 实体代表两个用户（或用户与客服）之间的私聊会话。
type Conversation struct {
	gorm.Model
	User1ID       uint64    `gorm:"index:idx_user1;not null;comment:用户1ID" json:"user1_id"`
	User2ID       uint64    `gorm:"index:idx_user2;not null;comment:用户2ID" json:"user2_id"`
	LastMessageID uint64    `gorm:"not null;comment:最后一条消息ID" json:"last_message_id"`
	LastMessage   string    `gorm:"type:varchar(255);comment:最后一条消息内容" json:"last_message"`
	LastMessageAt time.Time `gorm:"not null;comment:最后一条消息时间" json:"last_message_at"`
	UnreadCount1  int32     `gorm:"not null;default:0;comment:用户1未读数" json:"unread_count1"`
	UnreadCount2  int32     `gorm:"not null;default:0;comment:用户2未读数" json:"unread_count2"`
}

// ConversationMessage 实体代表会话中的一条消息。
type ConversationMessage struct {
	gorm.Model
	ConversationID uint64      `gorm:"index;not null;comment:会话ID" json:"conversation_id"`
	SenderID       uint64      `gorm:"index;not null;comment:发送者ID" json:"sender_id"`
	ReceiverID     uint64      `gorm:"index;not null;comment:接收者ID" json:"receiver_id"`
	Type           MessageType `gorm:"type:varchar(32);default:'TEXT';comment:消息类型" json:"type"`
	Content        string      `gorm:"type:text;not null;comment:内容" json:"content"`
	IsRead         bool        `gorm:"default:false;comment:是否已读" json:"is_read"`
	ReadAt         *time.Time  `gorm:"comment:阅读时间" json:"read_at"`
}

// NewConversation 创建会话。
func NewConversation(u1, u2 uint64) *Conversation {
	return &Conversation{
		User1ID:       u1,
		User2ID:       u2,
		LastMessageAt: time.Now(),
	}
}

// NewConversationMessage 创建消息。
func NewConversationMessage(convID, sender, receiver uint64, content string, msgType MessageType) *ConversationMessage {
	return &ConversationMessage{
		ConversationID: convID,
		SenderID:       sender,
		ReceiverID:     receiver,
		Content:        content,
		Type:           msgType,
		IsRead:         false,
	}
}

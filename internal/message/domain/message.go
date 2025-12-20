package domain

import (
	"time"

	"gorm.io/gorm"
)

// MessageType 定义了消息的类型。
type MessageType string

const (
	MessageTypeSystem  MessageType = "SYSTEM"
	MessageTypeOrder   MessageType = "ORDER"
	MessageTypeService MessageType = "SERVICE"
	MessageTypePromo   MessageType = "PROMO"
)

// MessageStatus 定义了消息的阅读状态。
type MessageStatus int8

const (
	MessageStatusUnread  MessageStatus = 0
	MessageStatusRead    MessageStatus = 1
	MessageStatusDeleted MessageStatus = 2
)

// Message 实体代表一条消息。
type Message struct {
	gorm.Model
	SenderID    uint64        `gorm:"not null;index;comment:发送者ID" json:"sender_id"`
	ReceiverID  uint64        `gorm:"not null;index;comment:接收者ID" json:"receiver_id"`
	MessageType MessageType   `gorm:"type:varchar(32);not null;comment:消息类型" json:"message_type"`
	Title       string        `gorm:"type:varchar(255);not null;comment:标题" json:"title"`
	Content     string        `gorm:"type:text;not null;comment:内容" json:"content"`
	Link        string        `gorm:"type:varchar(255);comment:链接" json:"link"`
	Status      MessageStatus `gorm:"type:tinyint;not null;default:0;comment:状态" json:"status"`
	ReadAt      *time.Time    `gorm:"comment:阅读时间" json:"read_at"`
}

// NewMessage 创建并返回一个新的 Message 实体实例。
func NewMessage(senderID, receiverID uint64, messageType MessageType, title, content, link string) *Message {
	return &Message{
		SenderID:    senderID,
		ReceiverID:  receiverID,
		MessageType: messageType,
		Title:       title,
		Content:     content,
		Link:        link,
		Status:      MessageStatusUnread,
	}
}

// MarkAsRead 标记消息为已读。
func (m *Message) MarkAsRead() {
	if m.Status == MessageStatusUnread {
		m.Status = MessageStatusRead
		now := time.Now()
		m.ReadAt = &now
	}
}

// Conversation 实体代表两个用户之间的会话。
type Conversation struct {
	gorm.Model
	User1ID       uint64    `gorm:"not null;index;comment:用户1ID" json:"user1_id"`
	User2ID       uint64    `gorm:"not null;index;comment:用户2ID" json:"user2_id"`
	LastMessageID uint64    `gorm:"not null;comment:最后一条消息ID" json:"last_message_id"`
	LastMessage   string    `gorm:"type:varchar(255);comment:最后一条消息内容" json:"last_message"`
	LastMessageAt time.Time `gorm:"not null;comment:最后一条消息时间" json:"last_message_at"`
	UnreadCount1  int32     `gorm:"not null;default:0;comment:用户1未读数" json:"unread_count1"`
	UnreadCount2  int32     `gorm:"not null;default:0;comment:用户2未读数" json:"unread_count2"`
}

// NewConversation 创建并返回一个新的 Conversation 实体实例。
func NewConversation(user1ID, user2ID uint64) *Conversation {
	return &Conversation{
		User1ID:      user1ID,
		User2ID:      user2ID,
		UnreadCount1: 0,
		UnreadCount2: 0,
	}
}

// UpdateLastMessage 更新会话的最后一条消息信息和未读计数。
func (c *Conversation) UpdateLastMessage(messageID uint64, message string, senderID uint64) {
	c.LastMessageID = messageID
	c.LastMessage = message
	c.LastMessageAt = time.Now()

	if senderID == c.User1ID {
		c.UnreadCount2++
	} else {
		c.UnreadCount1++
	}
}

// ClearUnreadCount 清空指定用户在会话中的未读消息计数。
func (c *Conversation) ClearUnreadCount(userID uint64) {
	switch userID {
	case c.User1ID:
		c.UnreadCount1 = 0
	case c.User2ID:
		c.UnreadCount2 = 0
	}
}

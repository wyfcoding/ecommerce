package entity

import (
	"time"

	"gorm.io/gorm" // 导入GORM库。
)

// MessageType 定义了消息的类型。
type MessageType string

const (
	MessageTypeSystem  MessageType = "SYSTEM"  // 系统消息：由系统自动发送。
	MessageTypeOrder   MessageType = "ORDER"   // 订单消息：与订单状态变更相关的通知。
	MessageTypeService MessageType = "SERVICE" // 客服消息：用户与客服之间的交流。
	MessageTypePromo   MessageType = "PROMO"   // 促销消息：营销推广信息。
)

// MessageStatus 定义了消息的阅读状态。
type MessageStatus int8

const (
	MessageStatusUnread  MessageStatus = 0 // 未读。
	MessageStatusRead    MessageStatus = 1 // 已读。
	MessageStatusDeleted MessageStatus = 2 // 已删除。
)

// Message 实体代表一条消息。
// 它包含了消息的发送者、接收者、类型、内容、链接和阅读状态等。
type Message struct {
	gorm.Model                // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	SenderID    uint64        `gorm:"not null;index;comment:发送者ID" json:"sender_id"`              // 发送者用户ID，索引字段。
	ReceiverID  uint64        `gorm:"not null;index;comment:接收者ID" json:"receiver_id"`            // 接收者用户ID，索引字段。
	MessageType MessageType   `gorm:"type:varchar(32);not null;comment:消息类型" json:"message_type"` // 消息类型。
	Title       string        `gorm:"type:varchar(255);not null;comment:标题" json:"title"`         // 消息标题。
	Content     string        `gorm:"type:text;not null;comment:内容" json:"content"`               // 消息内容。
	Link        string        `gorm:"type:varchar(255);comment:链接" json:"link"`                   // 消息关联的链接。
	Status      MessageStatus `gorm:"type:tinyint;not null;default:0;comment:状态" json:"status"`   // 消息状态，默认为未读。
	ReadAt      *time.Time    `gorm:"comment:阅读时间" json:"read_at"`                                // 消息被阅读的时间。
}

// NewMessage 创建并返回一个新的 Message 实体实例。
// senderID: 发送者用户ID。
// receiverID: 接收者用户ID。
// messageType: 消息类型。
// title: 消息标题。
// content: 消息内容。
// link: 消息关联的链接。
func NewMessage(senderID, receiverID uint64, messageType MessageType, title, content, link string) *Message {
	return &Message{
		SenderID:    senderID,
		ReceiverID:  receiverID,
		MessageType: messageType,
		Title:       title,
		Content:     content,
		Link:        link,
		Status:      MessageStatusUnread, // 新创建的消息默认为未读状态。
	}
}

// MarkAsRead 标记消息为已读。
func (m *Message) MarkAsRead() {
	if m.Status == MessageStatusUnread { // 只有未读消息才能标记为已读。
		m.Status = MessageStatusRead // 状态更新为已读。
		now := time.Now()
		m.ReadAt = &now // 记录阅读时间。
	}
}

// Conversation 实体代表两个用户之间的会话。
// 它记录了会话参与者、最后一条消息和未读计数等。
type Conversation struct {
	gorm.Model              // 嵌入gorm.Model。
	User1ID       uint64    `gorm:"not null;index;comment:用户1ID" json:"user1_id"`           // 会话参与者之一的用户ID，索引字段。
	User2ID       uint64    `gorm:"not null;index;comment:用户2ID" json:"user2_id"`           // 会话参与者之二的用户ID，索引字段。
	LastMessageID uint64    `gorm:"not null;comment:最后一条消息ID" json:"last_message_id"`       // 会话中最后一条消息的ID。
	LastMessage   string    `gorm:"type:varchar(255);comment:最后一条消息内容" json:"last_message"` // 会话中最后一条消息的内容摘要。
	LastMessageAt time.Time `gorm:"not null;comment:最后一条消息时间" json:"last_message_at"`       // 会话中最后一条消息的时间。
	UnreadCount1  int32     `gorm:"not null;default:0;comment:用户1未读数" json:"unread_count1"` // 用户1的未读消息数量。
	UnreadCount2  int32     `gorm:"not null;default:0;comment:用户2未读数" json:"unread_count2"` // 用户2的未读消息数量。
}

// NewConversation 创建并返回一个新的 Conversation 实体实例。
// user1ID: 会话参与者1的用户ID。
// user2ID: 会话参与者2的用户ID。
func NewConversation(user1ID, user2ID uint64) *Conversation {
	return &Conversation{
		User1ID:      user1ID,
		User2ID:      user2ID,
		UnreadCount1: 0, // 初始未读数为0。
		UnreadCount2: 0, // 初始未读数为0。
	}
}

// UpdateLastMessage 更新会话的最后一条消息信息和未读计数。
// messageID: 最后一条消息的ID。
// message: 最后一条消息的内容。
// senderID: 最后一条消息的发送者ID。
func (c *Conversation) UpdateLastMessage(messageID uint64, message string, senderID uint64) {
	c.LastMessageID = messageID  // 更新最后一条消息ID。
	c.LastMessage = message      // 更新最后一条消息内容。
	c.LastMessageAt = time.Now() // 更新最后一条消息时间为当前时间。

	// 根据发送者ID增加对应接收者的未读计数。
	if senderID == c.User1ID {
		c.UnreadCount2++
	} else {
		c.UnreadCount1++
	}
}

// ClearUnreadCount 清空指定用户在会话中的未读消息计数。
// userID: 待清空未读计数的`用户ID`。
func (c *Conversation) ClearUnreadCount(userID uint64) {
	switch userID {
	case c.User1ID:
		c.UnreadCount1 = 0
	case c.User2ID:
		c.UnreadCount2 = 0
	}
}

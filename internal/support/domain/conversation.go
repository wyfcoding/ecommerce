package domain

import "time"

// MessageType definitions are reused from ticket.go

// Conversation 实体代表两个用户（或用户与客服）之间的私聊会话。
type Conversation struct {
	ID            uint64    `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	User1ID       uint64    `json:"user1_id"`
	User2ID       uint64    `json:"user2_id"`
	LastMessageID uint64    `json:"last_message_id"`
	LastMessage   string    `json:"last_message"`
	LastMessageAt time.Time `json:"last_message_at"`
	UnreadCount1  int32     `json:"unread_count1"`
	UnreadCount2  int32     `json:"unread_count2"`
}

// ConversationMessage 实体代表会话中的一条消息。
type ConversationMessage struct {
	ID             uint64      `json:"id"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	ConversationID uint64      `json:"conversation_id"`
	SenderID       uint64      `json:"sender_id"`
	ReceiverID     uint64      `json:"receiver_id"`
	Type           MessageType `json:"type"`
	Content        string      `json:"content"`
	IsRead         bool        `json:"is_read"`
	ReadAt         *time.Time  `json:"read_at"`
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

package domain

import (
	"context"
)

// MessageRepository 是消息模块的仓储接口。
type MessageRepository interface {
	// Message
	SaveMessage(ctx context.Context, message *Message) error
	GetMessage(ctx context.Context, id uint64) (*Message, error)
	ListMessages(ctx context.Context, userID uint64, status *MessageStatus, offset, limit int) ([]*Message, int64, error)
	CountUnreadMessages(ctx context.Context, userID uint64) (int64, error)

	// Conversation
	SaveConversation(ctx context.Context, conversation *Conversation) error
	GetConversation(ctx context.Context, user1ID, user2ID uint64) (*Conversation, error)
	ListConversations(ctx context.Context, userID uint64, offset, limit int) ([]*Conversation, int64, error)
}

package repository

import (
	"context"
	"ecommerce/internal/message/domain/entity"
)

// MessageRepository 消息仓储接口
type MessageRepository interface {
	// 消息
	SaveMessage(ctx context.Context, message *entity.Message) error
	GetMessage(ctx context.Context, id uint64) (*entity.Message, error)
	ListMessages(ctx context.Context, userID uint64, status *entity.MessageStatus, offset, limit int) ([]*entity.Message, int64, error)
	CountUnreadMessages(ctx context.Context, userID uint64) (int64, error)

	// 会话
	SaveConversation(ctx context.Context, conversation *entity.Conversation) error
	GetConversation(ctx context.Context, user1ID, user2ID uint64) (*entity.Conversation, error)
	ListConversations(ctx context.Context, userID uint64, offset, limit int) ([]*entity.Conversation, int64, error)
}

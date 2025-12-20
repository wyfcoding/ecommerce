package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/message/domain"
)

// MessageQuery 处理消息的读操作。
type MessageQuery struct {
	repo domain.MessageRepository
}

// NewMessageQuery creates a new MessageQuery instance.
func NewMessageQuery(repo domain.MessageRepository) *MessageQuery {
	return &MessageQuery{
		repo: repo,
	}
}

// GetMessage 获取指定ID的消息详情。
func (q *MessageQuery) GetMessage(ctx context.Context, id uint64) (*domain.Message, error) {
	return q.repo.GetMessage(ctx, id)
}

// ListMessages 获取用户消息列表。
func (q *MessageQuery) ListMessages(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.Message, int64, error) {
	offset := (page - 1) * pageSize
	var msgStatus *domain.MessageStatus
	if status != nil {
		s := domain.MessageStatus(*status)
		msgStatus = &s
	}
	return q.repo.ListMessages(ctx, userID, msgStatus, offset, pageSize)
}

// GetUnreadCount 获取指定用户的未读消息数量。
func (q *MessageQuery) GetUnreadCount(ctx context.Context, userID uint64) (int64, error) {
	return q.repo.CountUnreadMessages(ctx, userID)
}

// ListConversations 获取用户会话列表。
func (q *MessageQuery) ListConversations(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.Conversation, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListConversations(ctx, userID, offset, pageSize)
}

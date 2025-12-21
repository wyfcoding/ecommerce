package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/message/domain"
)

// MessageService 作为消息操作的门面。
type MessageService struct {
	manager *MessageManager
	query   *MessageQuery
}

// NewMessageService 创建消息服务门面实例。
func NewMessageService(manager *MessageManager, query *MessageQuery) *MessageService {
	return &MessageService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

// SendMessage 发送一条站内信或系统消息。
func (s *MessageService) SendMessage(ctx context.Context, senderID, receiverID uint64, messageType domain.MessageType, title, content, link string) (*domain.Message, error) {
	return s.manager.SendMessage(ctx, senderID, receiverID, messageType, title, content, link)
}

// MarkAsRead 将指定消息标记为已读状态。
func (s *MessageService) MarkAsRead(ctx context.Context, id uint64, userID uint64) error {
	return s.manager.MarkAsRead(ctx, id, userID)
}

// --- 读操作（委托给 Query）---

// GetMessage 获取指定ID的消息详情。
func (s *MessageService) GetMessage(ctx context.Context, id uint64) (*domain.Message, error) {
	return s.query.GetMessage(ctx, id)
}

// ListMessages 获取用户接收到的消息列表。
func (s *MessageService) ListMessages(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.Message, int64, error) {
	return s.query.ListMessages(ctx, userID, status, page, pageSize)
}

// GetUnreadCount 获取指定用户的未读消息总数。
func (s *MessageService) GetUnreadCount(ctx context.Context, userID uint64) (int64, error) {
	return s.query.GetUnreadCount(ctx, userID)
}

// ListConversations 列出用户的会话记录列表。
func (s *MessageService) ListConversations(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.Conversation, int64, error) {
	return s.query.ListConversations(ctx, userID, page, pageSize)
}

package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/message/domain"
)

// MessageService acts as a facade for message operations.
type MessageService struct {
	manager *MessageManager
	query   *MessageQuery
}

// NewMessageService creates a new MessageService facade.
func NewMessageService(manager *MessageManager, query *MessageQuery) *MessageService {
	return &MessageService{
		manager: manager,
		query:   query,
	}
}

// --- Write Operations (Delegated to Manager) ---

func (s *MessageService) SendMessage(ctx context.Context, senderID, receiverID uint64, messageType domain.MessageType, title, content, link string) (*domain.Message, error) {
	return s.manager.SendMessage(ctx, senderID, receiverID, messageType, title, content, link)
}

func (s *MessageService) MarkAsRead(ctx context.Context, id uint64, userID uint64) error {
	return s.manager.MarkAsRead(ctx, id, userID)
}

// --- Read Operations (Delegated to Query) ---

func (s *MessageService) GetMessage(ctx context.Context, id uint64) (*domain.Message, error) {
	return s.query.GetMessage(ctx, id)
}

func (s *MessageService) ListMessages(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.Message, int64, error) {
	return s.query.ListMessages(ctx, userID, status, page, pageSize)
}

func (s *MessageService) GetUnreadCount(ctx context.Context, userID uint64) (int64, error) {
	return s.query.GetUnreadCount(ctx, userID)
}

func (s *MessageService) ListConversations(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.Conversation, int64, error) {
	return s.query.ListConversations(ctx, userID, page, pageSize)
}

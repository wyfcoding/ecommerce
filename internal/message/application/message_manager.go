package application

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/message/domain"

	"log/slog"
)

// MessageManager 处理消息的写操作。
type MessageManager struct {
	repo   domain.MessageRepository
	logger *slog.Logger
}

// NewMessageManager creates a new MessageManager instance.
func NewMessageManager(repo domain.MessageRepository, logger *slog.Logger) *MessageManager {
	return &MessageManager{
		repo:   repo,
		logger: logger,
	}
}

// SendMessage 发送一条消息。
func (m *MessageManager) SendMessage(ctx context.Context, senderID, receiverID uint64, messageType domain.MessageType, title, content, link string) (*domain.Message, error) {
	message := domain.NewMessage(senderID, receiverID, messageType, title, content, link)
	if err := m.repo.SaveMessage(ctx, message); err != nil {
		m.logger.Error("failed to save message", "error", err)
		return nil, err
	}

	if senderID > 0 && receiverID > 0 && messageType != domain.MessageTypeSystem {
		if err := m.updateConversation(ctx, senderID, receiverID, message); err != nil {
			m.logger.Warn("failed to update conversation", "error", err)
		}
	}

	return message, nil
}

// updateConversation 更新或创建会话记录。
func (m *MessageManager) updateConversation(ctx context.Context, senderID, receiverID uint64, message *domain.Message) error {
	conv, err := m.repo.GetConversation(ctx, senderID, receiverID)
	if err != nil {
		return err
	}

	if conv == nil {
		conv = domain.NewConversation(senderID, receiverID)
	}

	conv.UpdateLastMessage(uint64(message.ID), message.Content, senderID)
	return m.repo.SaveConversation(ctx, conv)
}

// MarkAsRead 标记指定消息为已读。
func (m *MessageManager) MarkAsRead(ctx context.Context, id uint64, userID uint64) error {
	message, err := m.repo.GetMessage(ctx, id)
	if err != nil {
		return err
	}
	if message == nil {
		return errors.New("message not found")
	}

	if message.ReceiverID != userID {
		return errors.New("permission denied")
	}

	message.MarkAsRead()
	return m.repo.SaveMessage(ctx, message)
}

package application

import (
	"context"
	"ecommerce/internal/message/domain/entity"
	"ecommerce/internal/message/domain/repository"
	"errors"

	"log/slog"
)

type MessageService struct {
	repo   repository.MessageRepository
	logger *slog.Logger
}

func NewMessageService(repo repository.MessageRepository, logger *slog.Logger) *MessageService {
	return &MessageService{
		repo:   repo,
		logger: logger,
	}
}

// SendMessage 发送消息
func (s *MessageService) SendMessage(ctx context.Context, senderID, receiverID uint64, messageType entity.MessageType, title, content, link string) (*entity.Message, error) {
	message := entity.NewMessage(senderID, receiverID, messageType, title, content, link)
	if err := s.repo.SaveMessage(ctx, message); err != nil {
		s.logger.Error("failed to save message", "error", err)
		return nil, err
	}

	// Update conversation if it's a user-to-user message (e.g., Service or custom type, assuming Service for now or just generic logic)
	// For simplicity, let's assume all messages might trigger conversation update if it's chat-like.
	// But MessageTypeSystem/Order/Promo usually don't have conversations.
	// Let's only update conversation for now if we decide to support chat.
	// Given the entity `Conversation` exists, we should probably support it.
	// Let's assume if senderID and receiverID are both non-zero and it's not system, we update conversation.

	if senderID > 0 && receiverID > 0 && messageType != entity.MessageTypeSystem {
		if err := s.updateConversation(ctx, senderID, receiverID, message); err != nil {
			s.logger.Warn("failed to update conversation", "error", err)
			// Don't fail the message send
		}
	}

	return message, nil
}

func (s *MessageService) updateConversation(ctx context.Context, senderID, receiverID uint64, message *entity.Message) error {
	conv, err := s.repo.GetConversation(ctx, senderID, receiverID)
	if err != nil {
		return err
	}

	if conv == nil {
		conv = entity.NewConversation(senderID, receiverID)
	}

	conv.UpdateLastMessage(uint64(message.ID), message.Content, senderID)
	return s.repo.SaveConversation(ctx, conv)
}

// GetMessage 获取消息
func (s *MessageService) GetMessage(ctx context.Context, id uint64) (*entity.Message, error) {
	return s.repo.GetMessage(ctx, id)
}

// MarkAsRead 标记已读
func (s *MessageService) MarkAsRead(ctx context.Context, id uint64, userID uint64) error {
	message, err := s.repo.GetMessage(ctx, id)
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
	return s.repo.SaveMessage(ctx, message)
}

// ListMessages 获取消息列表
func (s *MessageService) ListMessages(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*entity.Message, int64, error) {
	offset := (page - 1) * pageSize
	var msgStatus *entity.MessageStatus
	if status != nil {
		s := entity.MessageStatus(*status)
		msgStatus = &s
	}
	return s.repo.ListMessages(ctx, userID, msgStatus, offset, pageSize)
}

// GetUnreadCount 获取未读数
func (s *MessageService) GetUnreadCount(ctx context.Context, userID uint64) (int64, error) {
	return s.repo.CountUnreadMessages(ctx, userID)
}

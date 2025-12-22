package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/message/domain"

	"gorm.io/gorm"
)

type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository 创建并返回一个新的 messageRepository 实例。
func NewMessageRepository(db *gorm.DB) domain.MessageRepository {
	return &messageRepository{db: db}
}

// --- 消息管理 (Message methods) ---

func (r *messageRepository) SaveMessage(ctx context.Context, message *domain.Message) error {
	return r.db.WithContext(ctx).Save(message).Error
}

func (r *messageRepository) GetMessage(ctx context.Context, id uint64) (*domain.Message, error) {
	var message domain.Message
	if err := r.db.WithContext(ctx).First(&message, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果未找到则返回 nil, let application layer handle it? Or domain.ErrMessageNotFound?
			// 之前的实现返回 nil,nil。暂时保留或使用标准错误。
			// 除非有严格定义，否则按照接口预期返回 nil, nil。
		}
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) ListMessages(ctx context.Context, userID uint64, status *domain.MessageStatus, offset, limit int) ([]*domain.Message, int64, error) {
	var list []*domain.Message
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Message{}).Where("receiver_id = ?", userID)
	if status != nil {
		db = db.Where("status = ?", *status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *messageRepository) CountUnreadMessages(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Message{}).
		Where("receiver_id = ? AND status = ?", userID, domain.MessageStatusUnread).
		Count(&count).Error
	return count, err
}

// --- 会话管理 (Conversation methods) ---

func (r *messageRepository) SaveConversation(ctx context.Context, conversation *domain.Conversation) error {
	return r.db.WithContext(ctx).Save(conversation).Error
}

func (r *messageRepository) GetConversation(ctx context.Context, user1ID, user2ID uint64) (*domain.Conversation, error) {
	var conversation domain.Conversation
	err := r.db.WithContext(ctx).
		Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)", user1ID, user2ID, user2ID, user1ID).
		First(&conversation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &conversation, nil
}

func (r *messageRepository) ListConversations(ctx context.Context, userID uint64, offset, limit int) ([]*domain.Conversation, int64, error) {
	var list []*domain.Conversation
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Conversation{}).
		Where("user1_id = ? OR user2_id = ?", userID, userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("last_message_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

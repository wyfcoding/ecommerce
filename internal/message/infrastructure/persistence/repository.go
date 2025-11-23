package persistence

import (
	"context"
	"ecommerce/internal/message/domain/entity"
	"ecommerce/internal/message/domain/repository"
	"errors"

	"gorm.io/gorm"
)

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) repository.MessageRepository {
	return &messageRepository{db: db}
}

// 消息
func (r *messageRepository) SaveMessage(ctx context.Context, message *entity.Message) error {
	return r.db.WithContext(ctx).Save(message).Error
}

func (r *messageRepository) GetMessage(ctx context.Context, id uint64) (*entity.Message, error) {
	var message entity.Message
	if err := r.db.WithContext(ctx).First(&message, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) ListMessages(ctx context.Context, userID uint64, status *entity.MessageStatus, offset, limit int) ([]*entity.Message, int64, error) {
	var list []*entity.Message
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Message{}).Where("receiver_id = ?", userID)
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
	err := r.db.WithContext(ctx).Model(&entity.Message{}).
		Where("receiver_id = ? AND status = ?", userID, entity.MessageStatusUnread).
		Count(&count).Error
	return count, err
}

// 会话
func (r *messageRepository) SaveConversation(ctx context.Context, conversation *entity.Conversation) error {
	return r.db.WithContext(ctx).Save(conversation).Error
}

func (r *messageRepository) GetConversation(ctx context.Context, user1ID, user2ID uint64) (*entity.Conversation, error) {
	var conversation entity.Conversation
	// Ensure user1ID < user2ID for consistent querying if we enforce order, but here we just check both combinations or assume caller sorts
	// Let's assume caller sorts or we check both. For simplicity, let's query where (u1=a AND u2=b) OR (u1=b AND u2=a)
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

func (r *messageRepository) ListConversations(ctx context.Context, userID uint64, offset, limit int) ([]*entity.Conversation, int64, error) {
	var list []*entity.Conversation
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Conversation{}).
		Where("user1_id = ? OR user2_id = ?", userID, userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("last_message_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

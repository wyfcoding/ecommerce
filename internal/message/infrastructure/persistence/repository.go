package persistence

import (
	"context"
	"errors"
	"github.com/wyfcoding/ecommerce/internal/message/domain/entity"     // 导入消息模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/message/domain/repository" // 导入消息模块的领域仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type messageRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewMessageRepository 创建并返回一个新的 messageRepository 实例。
// db: GORM数据库连接实例。
func NewMessageRepository(db *gorm.DB) repository.MessageRepository {
	return &messageRepository{db: db}
}

// --- 消息管理 (Message methods) ---

// SaveMessage 将消息实体保存到数据库。
func (r *messageRepository) SaveMessage(ctx context.Context, message *entity.Message) error {
	return r.db.WithContext(ctx).Save(message).Error
}

// GetMessage 根据ID从数据库获取消息记录。
func (r *messageRepository) GetMessage(ctx context.Context, id uint64) (*entity.Message, error) {
	var message entity.Message
	if err := r.db.WithContext(ctx).First(&message, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &message, nil
}

// ListMessages 从数据库列出指定用户ID的所有消息记录，支持通过状态过滤和分页。
func (r *messageRepository) ListMessages(ctx context.Context, userID uint64, status *entity.MessageStatus, offset, limit int) ([]*entity.Message, int64, error) {
	var list []*entity.Message
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Message{}).Where("receiver_id = ?", userID)
	if status != nil { // 如果提供了状态，则按状态过滤。
		db = db.Where("status = ?", *status)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// CountUnreadMessages 统计指定用户ID的未读消息数量。
func (r *messageRepository) CountUnreadMessages(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	// 查询接收者为userID且状态为MessageStatusUnread的消息数量。
	err := r.db.WithContext(ctx).Model(&entity.Message{}).
		Where("receiver_id = ? AND status = ?", userID, entity.MessageStatusUnread).
		Count(&count).Error
	return count, err
}

// --- 会话管理 (Conversation methods) ---

// SaveConversation 将会话实体保存到数据库。
// 如果会话已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *messageRepository) SaveConversation(ctx context.Context, conversation *entity.Conversation) error {
	return r.db.WithContext(ctx).Save(conversation).Error
}

// GetConversation 根据两个用户ID获取会话记录。
// 为了保证查询的一致性，无论 user1ID 和 user2ID 的传入顺序如何，都会正确查找。
func (r *messageRepository) GetConversation(ctx context.Context, user1ID, user2ID uint64) (*entity.Conversation, error) {
	var conversation entity.Conversation
	// 查询条件：(user1_id = U1 AND user2_id = U2) OR (user1_id = U2 AND user2_id = U1)。
	err := r.db.WithContext(ctx).
		Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)", user1ID, user2ID, user2ID, user1ID).
		First(&conversation).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &conversation, nil
}

// ListConversations 从数据库列出指定用户ID参与的所有会话记录，支持分页。
func (r *messageRepository) ListConversations(ctx context.Context, userID uint64, offset, limit int) ([]*entity.Conversation, int64, error) {
	var list []*entity.Conversation
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Conversation{}).
		Where("user1_id = ? OR user2_id = ?", userID, userID) // 查找所有用户ID参与的会话。

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("last_message_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

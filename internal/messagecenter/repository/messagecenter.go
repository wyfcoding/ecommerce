package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"ecommerce/internal/messagecenter/model"
)

// MessageCenterRepo 消息中心仓储接口
type MessageCenterRepo interface {
	// 消息
	CreateMessage(ctx context.Context, message *model.Message) error
	GetMessageByID(ctx context.Context, id uint64) (*model.Message, error)
	ListMessages(ctx context.Context, messageType, status string, pageSize, pageNum int32) ([]*model.Message, int64, error)
	DeleteExpiredMessages(ctx context.Context) (int64, error)
	
	// 用户消息
	CreateUserMessage(ctx context.Context, userMessage *model.UserMessage) error
	BatchCreateUserMessages(ctx context.Context, userMessages []*model.UserMessage) error
	GetUserMessage(ctx context.Context, userID, messageID uint64) (*model.UserMessage, error)
	GetUserMessages(ctx context.Context, userID uint64, messageType, status string, pageSize, pageNum int32) ([]*model.UserMessage, int64, error)
	MarkAsRead(ctx context.Context, userID uint64, messageIDs []uint64, readAt time.Time) error
	MarkAllAsRead(ctx context.Context, userID uint64, messageType string, readAt time.Time) (int64, error)
	DeleteUserMessages(ctx context.Context, userID uint64, messageIDs []uint64) error
	GetUnreadCount(ctx context.Context, userID uint64) (int64, error)
	GetStatistics(ctx context.Context, userID uint64) (*model.MessageStatistics, error)
	
	// 消息模板
	CreateTemplate(ctx context.Context, template *model.MessageTemplate) error
	UpdateTemplate(ctx context.Context, template *model.MessageTemplate) error
	GetTemplateByCode(ctx context.Context, code string) (*model.MessageTemplate, error)
	ListTemplates(ctx context.Context) ([]*model.MessageTemplate, error)
	
	// 用户配置
	CreateUserConfig(ctx context.Context, config *model.MessageConfig) error
	UpdateUserConfig(ctx context.Context, config *model.MessageConfig) error
	GetUserConfig(ctx context.Context, userID uint64) (*model.MessageConfig, error)
}

type messageCenterRepo struct {
	db *gorm.DB
}

// NewMessageCenterRepo 创建消息中心仓储实例
func NewMessageCenterRepo(db *gorm.DB) MessageCenterRepo {
	return &messageCenterRepo{db: db}
}

// CreateMessage 创建消息
func (r *messageCenterRepo) CreateMessage(ctx context.Context, message *model.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

// GetMessageByID 根据ID获取消息
func (r *messageCenterRepo) GetMessageByID(ctx context.Context, id uint64) (*model.Message, error) {
	var message model.Message
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// ListMessages 获取消息列表
func (r *messageCenterRepo) ListMessages(ctx context.Context, messageType, status string, pageSize, pageNum int32) ([]*model.Message, int64, error) {
	var messages []*model.Message
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Message{})
	
	if messageType != "" {
		query = query.Where("type = ?", messageType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	err := query.Offset(int(offset)).Limit(int(pageSize)).Order("created_at DESC").Find(&messages).Error
	if err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// DeleteExpiredMessages 删除过期消息
func (r *messageCenterRepo) DeleteExpiredMessages(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).Where("expire_at IS NOT NULL AND expire_at < ?", time.Now()).Delete(&model.Message{})
	return result.RowsAffected, result.Error
}

// CreateUserMessage 创建用户消息
func (r *messageCenterRepo) CreateUserMessage(ctx context.Context, userMessage *model.UserMessage) error {
	return r.db.WithContext(ctx).Create(userMessage).Error
}

// BatchCreateUserMessages 批量创建用户消息
func (r *messageCenterRepo) BatchCreateUserMessages(ctx context.Context, userMessages []*model.UserMessage) error {
	return r.db.WithContext(ctx).CreateInBatches(userMessages, 100).Error
}

// GetUserMessage 获取用户消息
func (r *messageCenterRepo) GetUserMessage(ctx context.Context, userID, messageID uint64) (*model.UserMessage, error) {
	var userMessage model.UserMessage
	err := r.db.WithContext(ctx).
		Preload("Message").
		Where("user_id = ? AND message_id = ?", userID, messageID).
		First(&userMessage).Error
	if err != nil {
		return nil, err
	}
	return &userMessage, nil
}

// GetUserMessages 获取用户消息列表
func (r *messageCenterRepo) GetUserMessages(ctx context.Context, userID uint64, messageType, status string, pageSize, pageNum int32) ([]*model.UserMessage, int64, error) {
	var userMessages []*model.UserMessage
	var total int64

	query := r.db.WithContext(ctx).Model(&model.UserMessage{}).Where("user_id = ?", userID)
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 如果需要按消息类型筛选，需要join
	if messageType != "" {
		query = query.Joins("JOIN messages ON messages.id = user_messages.message_id").
			Where("messages.type = ?", messageType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	err := query.Preload("Message").
		Offset(int(offset)).
		Limit(int(pageSize)).
		Order("user_messages.created_at DESC").
		Find(&userMessages).Error
	if err != nil {
		return nil, 0, err
	}

	return userMessages, total, nil
}

// MarkAsRead 标记为已读
func (r *messageCenterRepo) MarkAsRead(ctx context.Context, userID uint64, messageIDs []uint64, readAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.UserMessage{}).
		Where("user_id = ? AND message_id IN ? AND status = ?", userID, messageIDs, model.MessageStatusUnread).
		Updates(map[string]interface{}{
			"status":  model.MessageStatusRead,
			"read_at": readAt,
		}).Error
}

// MarkAllAsRead 标记全部已读
func (r *messageCenterRepo) MarkAllAsRead(ctx context.Context, userID uint64, messageType string, readAt time.Time) (int64, error) {
	query := r.db.WithContext(ctx).Model(&model.UserMessage{}).
		Where("user_id = ? AND status = ?", userID, model.MessageStatusUnread)

	if messageType != "" {
		query = query.Joins("JOIN messages ON messages.id = user_messages.message_id").
			Where("messages.type = ?", messageType)
	}

	result := query.Updates(map[string]interface{}{
		"status":  model.MessageStatusRead,
		"read_at": readAt,
	})

	return result.RowsAffected, result.Error
}

// DeleteUserMessages 删除用户消息
func (r *messageCenterRepo) DeleteUserMessages(ctx context.Context, userID uint64, messageIDs []uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.UserMessage{}).
		Where("user_id = ? AND message_id IN ?", userID, messageIDs).
		Update("status", model.MessageStatusDeleted).Error
}

// GetUnreadCount 获取未读消息数
func (r *messageCenterRepo) GetUnreadCount(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserMessage{}).
		Where("user_id = ? AND status = ?", userID, model.MessageStatusUnread).
		Count(&count).Error
	return count, err
}

// GetStatistics 获取消息统计
func (r *messageCenterRepo) GetStatistics(ctx context.Context, userID uint64) (*model.MessageStatistics, error) {
	stats := &model.MessageStatistics{}

	// 总消息数
	r.db.WithContext(ctx).Model(&model.UserMessage{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalCount)

	// 未读消息数
	r.db.WithContext(ctx).Model(&model.UserMessage{}).
		Where("user_id = ? AND status = ?", userID, model.MessageStatusUnread).
		Count(&stats.UnreadCount)

	// 已读消息数
	r.db.WithContext(ctx).Model(&model.UserMessage{}).
		Where("user_id = ? AND status = ?", userID, model.MessageStatusRead).
		Count(&stats.ReadCount)

	// 按类型统计
	type TypeCount struct {
		Type  string
		Count int64
	}
	var typeCounts []TypeCount
	r.db.WithContext(ctx).Model(&model.UserMessage{}).
		Select("messages.type, COUNT(*) as count").
		Joins("JOIN messages ON messages.id = user_messages.message_id").
		Where("user_messages.user_id = ?", userID).
		Group("messages.type").
		Scan(&typeCounts)

	for _, tc := range typeCounts {
		switch tc.Type {
		case string(model.MessageTypeSystem):
			stats.SystemCount = tc.Count
		case string(model.MessageTypeOrder):
			stats.OrderCount = tc.Count
		case string(model.MessageTypePromotion):
			stats.PromotionCount = tc.Count
		case string(model.MessageTypeActivity):
			stats.ActivityCount = tc.Count
		case string(model.MessageTypeNotice):
			stats.NoticeCount = tc.Count
		case string(model.MessageTypeInteraction):
			stats.InteractionCount = tc.Count
		}
	}

	return stats, nil
}

// CreateTemplate 创建消息模板
func (r *messageCenterRepo) CreateTemplate(ctx context.Context, template *model.MessageTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

// UpdateTemplate 更新消息模板
func (r *messageCenterRepo) UpdateTemplate(ctx context.Context, template *model.MessageTemplate) error {
	return r.db.WithContext(ctx).Save(template).Error
}

// GetTemplateByCode 根据编码获取消息模板
func (r *messageCenterRepo) GetTemplateByCode(ctx context.Context, code string) (*model.MessageTemplate, error) {
	var template model.MessageTemplate
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// ListTemplates 获取消息模板列表
func (r *messageCenterRepo) ListTemplates(ctx context.Context) ([]*model.MessageTemplate, error) {
	var templates []*model.MessageTemplate
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// CreateUserConfig 创建用户配置
func (r *messageCenterRepo) CreateUserConfig(ctx context.Context, config *model.MessageConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// UpdateUserConfig 更新用户配置
func (r *messageCenterRepo) UpdateUserConfig(ctx context.Context, config *model.MessageConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// GetUserConfig 获取用户配置
func (r *messageCenterRepo) GetUserConfig(ctx context.Context, userID uint64) (*model.MessageConfig, error) {
	var config model.MessageConfig
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

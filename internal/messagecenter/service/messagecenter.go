package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ecommerce/internal/messagecenter/model"
	"ecommerce/internal/messagecenter/repository"
	"ecommerce/pkg/idgen"
)

var (
	ErrMessageNotFound   = errors.New("消息不存在")
	ErrTemplateNotFound  = errors.New("消息模板不存在")
	ErrTemplateNotActive = errors.New("消息模板未激活")
	ErrInvalidTargetType = errors.New("无效的目标类型")
)

// MessageCenterService 消息中心服务接口
type MessageCenterService interface {
	// 消息管理
	CreateMessage(ctx context.Context, message *model.Message) (*model.Message, error)
	GetMessage(ctx context.Context, id uint64) (*model.Message, error)
	SendMessage(ctx context.Context, message *model.Message, userIDs []uint64) error
	SendMessageByTemplate(ctx context.Context, templateCode string, userIDs []uint64, variables map[string]string) error
	
	// 用户消息
	GetUserMessages(ctx context.Context, userID uint64, messageType string, status string, pageSize, pageNum int32) ([]*model.UserMessage, int64, error)
	GetUserMessage(ctx context.Context, userID, messageID uint64) (*model.UserMessage, error)
	MarkAsRead(ctx context.Context, userID uint64, messageIDs []uint64) error
	MarkAllAsRead(ctx context.Context, userID uint64, messageType string) error
	DeleteUserMessage(ctx context.Context, userID uint64, messageIDs []uint64) error
	GetUnreadCount(ctx context.Context, userID uint64) (int64, error)
	GetStatistics(ctx context.Context, userID uint64) (*model.MessageStatistics, error)
	
	// 消息模板
	CreateTemplate(ctx context.Context, template *model.MessageTemplate) (*model.MessageTemplate, error)
	UpdateTemplate(ctx context.Context, template *model.MessageTemplate) (*model.MessageTemplate, error)
	GetTemplate(ctx context.Context, code string) (*model.MessageTemplate, error)
	ListTemplates(ctx context.Context) ([]*model.MessageTemplate, error)
	
	// 消息配置
	GetUserConfig(ctx context.Context, userID uint64) (*model.MessageConfig, error)
	UpdateUserConfig(ctx context.Context, config *model.MessageConfig) error
}

type messageCenterService struct {
	repo        repository.MessageCenterRepo
	redisClient *redis.Client
	logger      *zap.Logger
}

// NewMessageCenterService 创建消息中心服务实例
func NewMessageCenterService(
	repo repository.MessageCenterRepo,
	redisClient *redis.Client,
	logger *zap.Logger,
) MessageCenterService {
	return &messageCenterService{
		repo:        repo,
		redisClient: redisClient,
		logger:      logger,
	}
}

// CreateMessage 创建消息
func (s *messageCenterService) CreateMessage(ctx context.Context, message *model.Message) (*model.Message, error) {
	message.MessageNo = fmt.Sprintf("MSG%d", idgen.GenID())
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()

	if err := s.repo.CreateMessage(ctx, message); err != nil {
		s.logger.Error("创建消息失败", zap.Error(err))
		return nil, err
	}

	s.logger.Info("创建消息成功", zap.Uint64("messageID", message.ID))
	return message, nil
}

// GetMessage 获取消息详情
func (s *messageCenterService) GetMessage(ctx context.Context, id uint64) (*model.Message, error) {
	message, err := s.repo.GetMessageByID(ctx, id)
	if err != nil {
		return nil, ErrMessageNotFound
	}
	return message, nil
}

// SendMessage 发送消息
func (s *messageCenterService) SendMessage(ctx context.Context, message *model.Message, userIDs []uint64) error {
	// 1. 创建消息
	message, err := s.CreateMessage(ctx, message)
	if err != nil {
		return err
	}

	// 2. 根据目标类型处理
	var targetUserIDs []uint64
	switch message.TargetType {
	case "ALL":
		// 发送给所有用户（实际应该分批处理）
		// TODO: 实现分批发送
		targetUserIDs = userIDs
	case "USER":
		// 发送给指定用户
		targetUserIDs = userIDs
	case "GROUP":
		// 发送给用户组
		// TODO: 根据用户组获取用户列表
		targetUserIDs = userIDs
	default:
		return ErrInvalidTargetType
	}

	// 3. 创建用户消息关联
	userMessages := make([]*model.UserMessage, 0, len(targetUserIDs))
	now := time.Now()
	for _, userID := range targetUserIDs {
		// 检查用户消息配置
		if !s.shouldSendToUser(ctx, userID, message.Type) {
			continue
		}

		userMessages = append(userMessages, &model.UserMessage{
			UserID:    userID,
			MessageID: message.ID,
			Status:    model.MessageStatusUnread,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	if len(userMessages) == 0 {
		return nil
	}

	// 4. 批量创建用户消息
	if err := s.repo.BatchCreateUserMessages(ctx, userMessages); err != nil {
		s.logger.Error("批量创建用户消息失败", zap.Error(err))
		return err
	}

	// 5. 更新未读数缓存
	for _, userID := range targetUserIDs {
		s.incrementUnreadCount(ctx, userID)
	}

	// 6. 推送通知（异步）
	go s.pushNotification(context.Background(), message, targetUserIDs)

	s.logger.Info("发送消息成功",
		zap.Uint64("messageID", message.ID),
		zap.Int("userCount", len(targetUserIDs)))

	return nil
}

// SendMessageByTemplate 使用模板发送消息
func (s *messageCenterService) SendMessageByTemplate(ctx context.Context, templateCode string, userIDs []uint64, variables map[string]string) error {
	// 1. 获取模板
	template, err := s.repo.GetTemplateByCode(ctx, templateCode)
	if err != nil {
		return ErrTemplateNotFound
	}

	if !template.IsActive {
		return ErrTemplateNotActive
	}

	// 2. 渲染模板
	title := s.renderTemplate(template.Title, variables)
	content := s.renderTemplate(template.Content, variables)

	// 3. 创建消息
	message := &model.Message{
		Type:       template.Type,
		Priority:   model.MessagePriorityNormal,
		Title:      title,
		Content:    content,
		TargetType: "USER",
		PublishAt:  time.Now(),
	}

	// 4. 发送消息
	return s.SendMessage(ctx, message, userIDs)
}

// renderTemplate 渲染模板
func (s *messageCenterService) renderTemplate(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// shouldSendToUser 检查是否应该发送给用户
func (s *messageCenterService) shouldSendToUser(ctx context.Context, userID uint64, messageType model.MessageType) bool {
	config, err := s.repo.GetUserConfig(ctx, userID)
	if err != nil {
		// 如果没有配置，默认发送
		return true
	}

	// 检查消息类型开关
	switch messageType {
	case model.MessageTypeSystem:
		return config.SystemEnabled
	case model.MessageTypeOrder:
		return config.OrderEnabled
	case model.MessageTypePromotion:
		return config.PromotionEnabled
	case model.MessageTypeActivity:
		return config.ActivityEnabled
	case model.MessageTypeNotice:
		return config.NoticeEnabled
	case model.MessageTypeInteraction:
		return config.InteractionEnabled
	default:
		return true
	}
}

// GetUserMessages 获取用户消息列表
func (s *messageCenterService) GetUserMessages(ctx context.Context, userID uint64, messageType string, status string, pageSize, pageNum int32) ([]*model.UserMessage, int64, error) {
	messages, total, err := s.repo.GetUserMessages(ctx, userID, messageType, status, pageSize, pageNum)
	if err != nil {
		s.logger.Error("获取用户消息列表失败", zap.Error(err))
		return nil, 0, err
	}
	return messages, total, nil
}

// GetUserMessage 获取用户消息详情
func (s *messageCenterService) GetUserMessage(ctx context.Context, userID, messageID uint64) (*model.UserMessage, error) {
	userMessage, err := s.repo.GetUserMessage(ctx, userID, messageID)
	if err != nil {
		return nil, ErrMessageNotFound
	}

	// 自动标记为已读
	if userMessage.Status == model.MessageStatusUnread {
		s.MarkAsRead(ctx, userID, []uint64{messageID})
	}

	return userMessage, nil
}

// MarkAsRead 标记为已读
func (s *messageCenterService) MarkAsRead(ctx context.Context, userID uint64, messageIDs []uint64) error {
	now := time.Now()
	if err := s.repo.MarkAsRead(ctx, userID, messageIDs, now); err != nil {
		s.logger.Error("标记消息已读失败", zap.Error(err))
		return err
	}

	// 更新未读数缓存
	s.decrementUnreadCount(ctx, userID, int64(len(messageIDs)))

	return nil
}

// MarkAllAsRead 标记全部已读
func (s *messageCenterService) MarkAllAsRead(ctx context.Context, userID uint64, messageType string) error {
	now := time.Now()
	count, err := s.repo.MarkAllAsRead(ctx, userID, messageType, now)
	if err != nil {
		s.logger.Error("标记全部已读失败", zap.Error(err))
		return err
	}

	// 更新未读数缓存
	if messageType == "" {
		// 清空未读数
		s.clearUnreadCount(ctx, userID)
	} else {
		s.decrementUnreadCount(ctx, userID, count)
	}

	return nil
}

// DeleteUserMessage 删除用户消息
func (s *messageCenterService) DeleteUserMessage(ctx context.Context, userID uint64, messageIDs []uint64) error {
	if err := s.repo.DeleteUserMessages(ctx, userID, messageIDs); err != nil {
		s.logger.Error("删除用户消息失败", zap.Error(err))
		return err
	}

	return nil
}

// GetUnreadCount 获取未读消息数
func (s *messageCenterService) GetUnreadCount(ctx context.Context, userID uint64) (int64, error) {
	// 先从缓存获取
	cacheKey := fmt.Sprintf("message:unread:%d", userID)
	count, err := s.redisClient.Get(ctx, cacheKey).Int64()
	if err == nil {
		return count, nil
	}

	// 缓存未命中，从数据库查询
	count, err = s.repo.GetUnreadCount(ctx, userID)
	if err != nil {
		return 0, err
	}

	// 写入缓存
	s.redisClient.Set(ctx, cacheKey, count, 24*time.Hour)

	return count, nil
}

// GetStatistics 获取消息统计
func (s *messageCenterService) GetStatistics(ctx context.Context, userID uint64) (*model.MessageStatistics, error) {
	stats, err := s.repo.GetStatistics(ctx, userID)
	if err != nil {
		s.logger.Error("获取消息统计失败", zap.Error(err))
		return nil, err
	}
	return stats, nil
}

// CreateTemplate 创建消息模板
func (s *messageCenterService) CreateTemplate(ctx context.Context, template *model.MessageTemplate) (*model.MessageTemplate, error) {
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	if err := s.repo.CreateTemplate(ctx, template); err != nil {
		s.logger.Error("创建消息模板失败", zap.Error(err))
		return nil, err
	}

	return template, nil
}

// UpdateTemplate 更新消息模板
func (s *messageCenterService) UpdateTemplate(ctx context.Context, template *model.MessageTemplate) (*model.MessageTemplate, error) {
	template.UpdatedAt = time.Now()

	if err := s.repo.UpdateTemplate(ctx, template); err != nil {
		s.logger.Error("更新消息模板失败", zap.Error(err))
		return nil, err
	}

	return template, nil
}

// GetTemplate 获取消息模板
func (s *messageCenterService) GetTemplate(ctx context.Context, code string) (*model.MessageTemplate, error) {
	template, err := s.repo.GetTemplateByCode(ctx, code)
	if err != nil {
		return nil, ErrTemplateNotFound
	}
	return template, nil
}

// ListTemplates 获取模板列表
func (s *messageCenterService) ListTemplates(ctx context.Context) ([]*model.MessageTemplate, error) {
	templates, err := s.repo.ListTemplates(ctx)
	if err != nil {
		s.logger.Error("获取模板列表失败", zap.Error(err))
		return nil, err
	}
	return templates, nil
}

// GetUserConfig 获取用户消息配置
func (s *messageCenterService) GetUserConfig(ctx context.Context, userID uint64) (*model.MessageConfig, error) {
	config, err := s.repo.GetUserConfig(ctx, userID)
	if err != nil {
		// 如果不存在，创建默认配置
		config = &model.MessageConfig{
			UserID:             userID,
			SystemEnabled:      true,
			OrderEnabled:       true,
			PromotionEnabled:   true,
			ActivityEnabled:    true,
			NoticeEnabled:      true,
			InteractionEnabled: true,
			PushEnabled:        true,
			EmailEnabled:       false,
			SMSEnabled:         false,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}
		s.repo.CreateUserConfig(ctx, config)
	}
	return config, nil
}

// UpdateUserConfig 更新用户消息配置
func (s *messageCenterService) UpdateUserConfig(ctx context.Context, config *model.MessageConfig) error {
	config.UpdatedAt = time.Now()

	if err := s.repo.UpdateUserConfig(ctx, config); err != nil {
		s.logger.Error("更新用户消息配置失败", zap.Error(err))
		return err
	}

	return nil
}

// incrementUnreadCount 增加未读数
func (s *messageCenterService) incrementUnreadCount(ctx context.Context, userID uint64) {
	cacheKey := fmt.Sprintf("message:unread:%d", userID)
	s.redisClient.Incr(ctx, cacheKey)
	s.redisClient.Expire(ctx, cacheKey, 24*time.Hour)
}

// decrementUnreadCount 减少未读数
func (s *messageCenterService) decrementUnreadCount(ctx context.Context, userID uint64, count int64) {
	cacheKey := fmt.Sprintf("message:unread:%d", userID)
	s.redisClient.DecrBy(ctx, cacheKey, count)
}

// clearUnreadCount 清空未读数
func (s *messageCenterService) clearUnreadCount(ctx context.Context, userID uint64) {
	cacheKey := fmt.Sprintf("message:unread:%d", userID)
	s.redisClient.Del(ctx, cacheKey)
}

// pushNotification 推送通知（异步）
func (s *messageCenterService) pushNotification(ctx context.Context, message *model.Message, userIDs []uint64) {
	// TODO: 实现推送通知
	// 1. 检查用户推送配置
	// 2. 检查免打扰时段
	// 3. 调用推送服务（极光推送、个推等）
	// 4. 记录推送日志

	s.logger.Info("推送通知",
		zap.Uint64("messageID", message.ID),
		zap.Int("userCount", len(userIDs)))
}

// SendSystemMessage 发送系统消息（便捷方法）
func (s *messageCenterService) SendSystemMessage(ctx context.Context, title, content string, userIDs []uint64) error {
	message := &model.Message{
		Type:       model.MessageTypeSystem,
		Priority:   model.MessagePriorityNormal,
		Title:      title,
		Content:    content,
		TargetType: "USER",
		PublishAt:  time.Now(),
	}
	return s.SendMessage(ctx, message, userIDs)
}

// SendOrderMessage 发送订单消息（便捷方法）
func (s *messageCenterService) SendOrderMessage(ctx context.Context, userID uint64, orderNo, title, content string) error {
	message := &model.Message{
		Type:       model.MessageTypeOrder,
		Priority:   model.MessagePriorityHigh,
		Title:      title,
		Content:    content,
		LinkURL:    fmt.Sprintf("/order/detail?orderNo=%s", orderNo),
		LinkType:   "ORDER",
		TargetType: "USER",
		PublishAt:  time.Now(),
	}
	return s.SendMessage(ctx, message, []uint64{userID})
}

// BroadcastMessage 广播消息
func (s *messageCenterService) BroadcastMessage(ctx context.Context, message *model.Message) error {
	message.TargetType = "ALL"
	
	// TODO: 实现分批广播
	// 1. 获取所有用户ID（分批）
	// 2. 分批发送消息
	// 3. 记录发送进度
	
	return s.SendMessage(ctx, message, []uint64{})
}

// CleanExpiredMessages 清理过期消息（定时任务）
func (s *messageCenterService) CleanExpiredMessages(ctx context.Context) (int64, error) {
	count, err := s.repo.DeleteExpiredMessages(ctx)
	if err != nil {
		s.logger.Error("清理过期消息失败", zap.Error(err))
		return 0, err
	}

	s.logger.Info("清理过期消息成功", zap.Int64("count", count))
	return count, nil
}

// ExportUserMessages 导出用户消息（用于数据备份）
func (s *messageCenterService) ExportUserMessages(ctx context.Context, userID uint64, startTime, endTime time.Time) ([]byte, error) {
	messages, _, err := s.repo.GetUserMessages(ctx, userID, "", "", 10000, 1)
	if err != nil {
		return nil, err
	}

	// 转换为JSON
	data, err := json.Marshal(messages)
	if err != nil {
		return nil, err
	}

	return data, nil
}

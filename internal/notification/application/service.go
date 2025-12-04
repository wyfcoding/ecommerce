package application

import (
	"context"
	"errors"  // 导入标准错误处理库。
	"strings" // 导入字符串操作库。

	"github.com/wyfcoding/ecommerce/internal/notification/domain/entity"     // 导入通知领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/notification/domain/repository" // 导入通知领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// NotificationService 结构体定义了通知管理相关的应用服务。
// 它协调领域层和基础设施层，处理通知的发送、读取状态管理和通知模板管理等业务逻辑。
type NotificationService struct {
	repo   repository.NotificationRepository // 依赖NotificationRepository接口，用于数据持久化操作。
	logger *slog.Logger                      // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewNotificationService 创建并返回一个新的 NotificationService 实例。
func NewNotificationService(repo repository.NotificationRepository, logger *slog.Logger) *NotificationService {
	return &NotificationService{
		repo:   repo,
		logger: logger,
	}
}

// SendNotification 发送一条通知。
// ctx: 上下文。
// userID: 接收通知的用户ID。
// notifType: 通知类型。
// channel: 通知渠道。
// title: 通知标题。
// content: 通知内容。
// data: 附加数据（例如，JSON格式）。
// 返回发送成功的Notification实体和可能发生的错误。
func (s *NotificationService) SendNotification(ctx context.Context, userID uint64, notifType entity.NotificationType, channel entity.NotificationChannel, title, content string, data map[string]interface{}) (*entity.Notification, error) {
	notification := entity.NewNotification(userID, notifType, channel, title, content, data) // 创建Notification实体。
	// 通过仓储接口保存通知。
	if err := s.repo.SaveNotification(ctx, notification); err != nil {
		s.logger.Error("failed to save notification", "error", err)
		return nil, err
	}

	// TODO: 在实际系统中，此处应触发实际的通知发送机制（例如，调用推送服务、邮件服务或短信服务）。
	// 当前实现仅模拟发送，并记录日志。
	s.logger.Info("Notification sent (simulated)",
		"user_id", userID,
		"type", string(notifType),
		"channel", string(channel))

	return notification, nil
}

// SendNotificationByTemplate 使用指定的模板发送通知。
// ctx: 上下文。
// userID: 接收通知的用户ID。
// templateCode: 模板编码。
// variables: 模板中待替换的变量。
// data: 附加数据。
// 返回发送成功的Notification实体和可能发生的错误。
func (s *NotificationService) SendNotificationByTemplate(ctx context.Context, userID uint64, templateCode string, variables map[string]string, data map[string]interface{}) (*entity.Notification, error) {
	// 获取通知模板。
	template, err := s.repo.GetTemplateByCode(ctx, templateCode)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, errors.New("template not found")
	}
	if !template.Enabled {
		return nil, errors.New("template disabled")
	}

	// 替换模板中的变量。
	title := template.Title
	content := template.Content
	for key, val := range variables {
		title = strings.ReplaceAll(title, "{{"+key+"}}", val)
		content = strings.ReplaceAll(content, "{{"+key+"}}", val)
	}

	// 调用 SendNotification 方法发送实际通知。
	return s.SendNotification(ctx, userID, template.NotifType, template.Channel, title, content, data)
}

// GetNotification 获取指定ID的通知详情。
// ctx: 上下文。
// id: 通知ID。
// 返回Notification实体和可能发生的错误。
func (s *NotificationService) GetNotification(ctx context.Context, id uint64) (*entity.Notification, error) {
	return s.repo.GetNotification(ctx, id)
}

// MarkAsRead 标记指定通知为已读。
// ctx: 上下文。
// id: 通知ID。
// userID: 操作用户ID，用于权限验证（确保只有接收者可以标记为已读）。
// 返回可能发生的错误。
func (s *NotificationService) MarkAsRead(ctx context.Context, id uint64, userID uint64) error {
	notification, err := s.repo.GetNotification(ctx, id)
	if err != nil {
		return err
	}
	if notification == nil {
		return errors.New("notification not found")
	}

	// 权限验证：确保当前用户是通知的接收者。
	if notification.UserID != userID {
		return errors.New("permission denied")
	}

	// 调用实体方法标记消息为已读。
	notification.MarkAsRead()
	// 保存更新后的通知。
	return s.repo.SaveNotification(ctx, notification)
}

// ListNotifications 获取用户通知列表。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// status: 通知状态（例如，未读、已读）。
// page, pageSize: 分页参数。
// 返回通知列表、总数和可能发生的错误。
func (s *NotificationService) ListNotifications(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*entity.Notification, int64, error) {
	offset := (page - 1) * pageSize
	var notifStatus *entity.NotificationStatus
	if status != nil {
		s := entity.NotificationStatus(*status)
		notifStatus = &s
	}
	return s.repo.ListNotifications(ctx, userID, notifStatus, offset, pageSize)
}

// GetUnreadCount 获取指定用户的未读通知数量。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// 返回未读通知数量和可能发生的错误。
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uint64) (int64, error) {
	return s.repo.CountUnreadNotifications(ctx, userID)
}

// CreateTemplate 创建一个通知模板。
// ctx: 上下文。
// template: 待创建的NotificationTemplate实体。
// 返回可能发生的错误。
func (s *NotificationService) CreateTemplate(ctx context.Context, template *entity.NotificationTemplate) error {
	return s.repo.SaveTemplate(ctx, template)
}

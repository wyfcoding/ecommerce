package application

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/notification/domain"
)

// NotificationManager 处理通知的写操作（发送、标记已读、模板管理）。
type NotificationManager struct {
	repo   domain.NotificationRepository
	logger *slog.Logger
}

// NewNotificationManager 负责处理 NewNotification 相关的写操作和业务逻辑。
func NewNotificationManager(repo domain.NotificationRepository, logger *slog.Logger) *NotificationManager {
	return &NotificationManager{
		repo:   repo,
		logger: logger,
	}
}

// SendNotification 发送一条通知。
func (m *NotificationManager) SendNotification(ctx context.Context, userID uint64, notifType domain.NotificationType, channel domain.NotificationChannel, title, content string, data map[string]any) (*domain.Notification, error) {
	notification := domain.NewNotification(userID, notifType, channel, title, content, data)
	if err := m.repo.SaveNotification(ctx, notification); err != nil {
		m.logger.Error("failed to save notification", "error", err)
		return nil, err
	}

	// Simulated send
	m.logger.Info("Notification sent (simulated)",
		"user_id", userID,
		"type", string(notifType),
		"channel", string(channel))

	return notification, nil
}

// SendNotificationByTemplate 使用指定的模板发送通知。
func (m *NotificationManager) SendNotificationByTemplate(ctx context.Context, userID uint64, templateCode string, variables map[string]string, data map[string]any) (*domain.Notification, error) {
	template, err := m.repo.GetTemplateByCode(ctx, templateCode)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, errors.New("template not found")
	}
	if !template.Enabled {
		return nil, errors.New("template disabled")
	}

	title := template.Title
	content := template.Content
	for key, val := range variables {
		title = strings.ReplaceAll(title, "{{"+key+"}}", val)
		content = strings.ReplaceAll(content, "{{"+key+"}}", val)
	}

	return m.SendNotification(ctx, userID, template.NotifType, template.Channel, title, content, data)
}

// MarkAsRead 标记指定通知为已读。
func (m *NotificationManager) MarkAsRead(ctx context.Context, id uint64, userID uint64) error {
	notification, err := m.repo.GetNotification(ctx, id)
	if err != nil {
		return err
	}
	if notification == nil {
		return errors.New("notification not found")
	}

	if notification.UserID != userID {
		return errors.New("permission denied")
	}

	notification.MarkAsRead()
	return m.repo.SaveNotification(ctx, notification)
}

// DeleteNotification 删除一条通知。
func (m *NotificationManager) DeleteNotification(ctx context.Context, id uint64) error {
	if err := m.repo.DeleteNotification(ctx, id); err != nil {
		m.logger.ErrorContext(ctx, "failed to delete notification", "id", id, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "notification deleted successfully", "id", id)
	return nil
}

// CreateTemplate 创建一个通知模板。
func (m *NotificationManager) CreateTemplate(ctx context.Context, template *domain.NotificationTemplate) error {
	return m.repo.SaveTemplate(ctx, template)
}

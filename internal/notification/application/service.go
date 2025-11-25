package application

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/notification/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/notification/domain/repository"
	"errors"
	"strings"

	"log/slog"
)

type NotificationService struct {
	repo   repository.NotificationRepository
	logger *slog.Logger
}

func NewNotificationService(repo repository.NotificationRepository, logger *slog.Logger) *NotificationService {
	return &NotificationService{
		repo:   repo,
		logger: logger,
	}
}

// SendNotification 发送通知
func (s *NotificationService) SendNotification(ctx context.Context, userID uint64, notifType entity.NotificationType, channel entity.NotificationChannel, title, content string, data map[string]interface{}) (*entity.Notification, error) {
	notification := entity.NewNotification(userID, notifType, channel, title, content, data)
	if err := s.repo.SaveNotification(ctx, notification); err != nil {
		s.logger.Error("failed to save notification", "error", err)
		return nil, err
	}

	// In a real system, we would also trigger the actual sending (e.g., push notification, email, SMS) here
	// using an external provider or a message queue.
	s.logger.Info("Notification sent (simulated)",
		"user_id", userID,
		"type", string(notifType),
		"channel", string(channel))

	return notification, nil
}

// SendNotificationByTemplate 使用模板发送通知
func (s *NotificationService) SendNotificationByTemplate(ctx context.Context, userID uint64, templateCode string, variables map[string]string, data map[string]interface{}) (*entity.Notification, error) {
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

	// Replace variables
	title := template.Title
	content := template.Content
	for key, val := range variables {
		title = strings.ReplaceAll(title, "{{"+key+"}}", val)
		content = strings.ReplaceAll(content, "{{"+key+"}}", val)
	}

	return s.SendNotification(ctx, userID, template.NotifType, template.Channel, title, content, data)
}

// GetNotification 获取通知
func (s *NotificationService) GetNotification(ctx context.Context, id uint64) (*entity.Notification, error) {
	return s.repo.GetNotification(ctx, id)
}

// MarkAsRead 标记已读
func (s *NotificationService) MarkAsRead(ctx context.Context, id uint64, userID uint64) error {
	notification, err := s.repo.GetNotification(ctx, id)
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
	return s.repo.SaveNotification(ctx, notification)
}

// ListNotifications 获取通知列表
func (s *NotificationService) ListNotifications(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*entity.Notification, int64, error) {
	offset := (page - 1) * pageSize
	var notifStatus *entity.NotificationStatus
	if status != nil {
		s := entity.NotificationStatus(*status)
		notifStatus = &s
	}
	return s.repo.ListNotifications(ctx, userID, notifStatus, offset, pageSize)
}

// GetUnreadCount 获取未读数
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uint64) (int64, error) {
	return s.repo.CountUnreadNotifications(ctx, userID)
}

// CreateTemplate 创建模板
func (s *NotificationService) CreateTemplate(ctx context.Context, template *entity.NotificationTemplate) error {
	return s.repo.SaveTemplate(ctx, template)
}

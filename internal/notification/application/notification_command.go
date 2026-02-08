package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"github.com/wyfcoding/ecommerce/internal/notification/domain"
	"github.com/wyfcoding/pkg/messagequeue"
	"github.com/wyfcoding/pkg/server"
)

// NotificationCommandService 处理通知的写操作（发送、标记已读、模板管理）。
type NotificationCommandService struct {
	repo             domain.NotificationRepository
	templateReadRepo domain.NotificationTemplateReadRepository
	publisher        messagequeue.EventPublisher
	emailSender      domain.Sender
	smsSender        domain.Sender
	webhookSender    domain.Sender
	websocketMgr     *server.WSManager
	logger           *slog.Logger
}

// NewNotificationCommandService 创建写服务实例。
func NewNotificationCommandService(
	repo domain.NotificationRepository,
	templateReadRepo domain.NotificationTemplateReadRepository,
	publisher messagequeue.EventPublisher,
	emailSender, smsSender, webhookSender domain.Sender,
	websocketMgr *server.WSManager,
	logger *slog.Logger,
) *NotificationCommandService {
	return &NotificationCommandService{
		repo:             repo,
		templateReadRepo: templateReadRepo,
		publisher:        publisher,
		emailSender:      emailSender,
		smsSender:        smsSender,
		webhookSender:    webhookSender,
		websocketMgr:     websocketMgr,
		logger:           logger,
	}
}

// SendNotification 发送一条通知。
func (m *NotificationCommandService) SendNotification(ctx context.Context, userID uint64, notifType domain.NotificationType, channel domain.NotificationChannel, title, content string, data map[string]any) (*domain.Notification, error) {
	notification := domain.NewNotification(userID, notifType, channel, title, content, data)
	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveNotificationInTx(ctx, tx, notification); err != nil {
			m.logger.Error("failed to save notification", "error", err)
			return err
		}
		if m.publisher != nil {
			event := &domain.NotificationCreatedEvent{
				NotificationID: notification.ID,
				UserID:         userID,
				Timestamp:      time.Now(),
			}
			if err := m.publisher.PublishInTx(ctx, tx, domain.NotificationCreatedEventType, fmt.Sprintf("%d", notification.ID), event); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// 真实化执行：根据渠道调用相应的发送器
	var sendErr error
	target := "target_user_identity"
	if data != nil {
		if t, ok := data["target"].(string); ok {
			target = t
		}
	}

	switch channel {
	case domain.NotificationChannelEmail:
		if m.emailSender != nil {
			sendErr = m.emailSender.Send(ctx, target, title, content)
		}
	case domain.NotificationChannelSMS:
		if m.smsSender != nil {
			sendErr = m.smsSender.Send(ctx, target, title, content)
		}
	case domain.NotificationChannelApp:
		m.logger.Info("in-app notification persisted", "user_id", userID)
		// 通过 Websocket 推送给在线用户
		userIDStr := strconv.FormatUint(userID, 10)
		if m.websocketMgr != nil && m.websocketMgr.IsUserOnline(userIDStr) {
			payload, _ := json.Marshal(map[string]any{
				"type":    "notification",
				"id":      notification.ID,
				"title":   title,
				"content": content,
				"data":    data,
			})
			m.websocketMgr.SendToUserRaw(userIDStr, payload)
			m.logger.Info("notification pushed via websocket", "user_id", userID)
		}
	case domain.NotificationChannelWebhook:
		if m.webhookSender != nil {
			sendErr = m.webhookSender.Send(ctx, target, title, content)
		}
	}

	if sendErr != nil {
		m.logger.Error("failed to send notification via channel", "channel", channel, "error", sendErr)
	} else {
		m.logger.Info("notification sent successfully", "user_id", userID, "channel", channel)
	}

	return notification, nil
}

// SendNotificationByTemplate 使用指定的模板发送通知。
func (m *NotificationCommandService) SendNotificationByTemplate(ctx context.Context, userID uint64, templateCode string, variables map[string]string, data map[string]any) (*domain.Notification, error) {
	var template *domain.NotificationTemplate
	if m.templateReadRepo != nil {
		if cached, err := m.templateReadRepo.GetByCode(ctx, templateCode); err == nil && cached != nil {
			template = cached
		}
	}
	if template == nil {
		var err error
		template, err = m.repo.GetTemplateByCode(ctx, templateCode)
		if err != nil {
			return nil, err
		}
		if template != nil && m.templateReadRepo != nil {
			_ = m.templateReadRepo.Save(ctx, template)
		}
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
func (m *NotificationCommandService) MarkAsRead(ctx context.Context, id uint64, userID uint64) error {
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
	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveNotificationInTx(ctx, tx, notification); err != nil {
			return err
		}
		if m.publisher != nil {
			event := &domain.NotificationReadEvent{
				NotificationID: id,
				UserID:         userID,
				Timestamp:      time.Now(),
			}
			if err := m.publisher.PublishInTx(ctx, tx, domain.NotificationReadEventType, fmt.Sprintf("%d", id), event); err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteNotification 删除一条通知。
func (m *NotificationCommandService) DeleteNotification(ctx context.Context, id uint64) error {
	notification, err := m.repo.GetNotification(ctx, id)
	if err != nil {
		return err
	}
	if notification == nil {
		return nil
	}

	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.DeleteNotificationInTx(ctx, tx, id); err != nil {
			m.logger.ErrorContext(ctx, "failed to delete notification", "id", id, "error", err)
			return err
		}
		if m.publisher != nil {
			event := &domain.NotificationDeletedEvent{
				NotificationID: id,
				UserID:         notification.UserID,
				Timestamp:      time.Now(),
			}
			if err := m.publisher.PublishInTx(ctx, tx, domain.NotificationDeletedEventType, fmt.Sprintf("%d", id), event); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	m.logger.InfoContext(ctx, "notification deleted successfully", "id", id)
	return nil
}

// CreateTemplate 创建一个通知模板。
func (m *NotificationCommandService) CreateTemplate(ctx context.Context, template *domain.NotificationTemplate) error {
	if template == nil {
		return nil
	}
	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveTemplateInTx(ctx, tx, template); err != nil {
			return err
		}
		if m.publisher != nil {
			event := &domain.NotificationTemplateCreatedEvent{
				TemplateID: template.ID,
				Code:       template.Code,
				Timestamp:  time.Now(),
			}
			if err := m.publisher.PublishInTx(ctx, tx, domain.NotificationTemplateCreatedEventType, template.Code, event); err != nil {
				return err
			}
		}
		return nil
	})
}

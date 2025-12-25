package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/notification/domain"
)

// Notification 结构体定义了通知管理相关的应用服务（外观模式）。
// 它将业务逻辑委托给 NotificationManager 和 NotificationQuery 处理。
type Notification struct {
	manager *NotificationManager
	query   *NotificationQuery
}

// NewNotification 创建并返回一个新的 Notification 实例。
func NewNotification(manager *NotificationManager, query *NotificationQuery) *Notification {
	return &Notification{
		manager: manager,
		query:   query,
	}
}

// SendNotification 发送一条通知。
func (s *Notification) SendNotification(ctx context.Context, userID uint64, notifType domain.NotificationType, channel domain.NotificationChannel, title, content string, data map[string]any) (*domain.Notification, error) {
	return s.manager.SendNotification(ctx, userID, notifType, channel, title, content, data)
}

// SendNotificationByTemplate 使用指定的模板发送通知。
func (s *Notification) SendNotificationByTemplate(ctx context.Context, userID uint64, templateCode string, variables map[string]string, data map[string]any) (*domain.Notification, error) {
	return s.manager.SendNotificationByTemplate(ctx, userID, templateCode, variables, data)
}

// GetNotification 获取指定ID的通知详情。
func (s *Notification) GetNotification(ctx context.Context, id uint64) (*domain.Notification, error) {
	return s.query.GetNotification(ctx, id)
}

// MarkAsRead 标记指定通知为已读。
func (s *Notification) MarkAsRead(ctx context.Context, id uint64, userID uint64) error {
	return s.manager.MarkAsRead(ctx, id, userID)
}

// ListNotifications 获取指定用户的通知列表。
func (s *Notification) ListNotifications(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.Notification, int64, error) {
	return s.query.ListNotifications(ctx, userID, status, page, pageSize)
}

// GetUnreadCount 获取指定用户的未读通知数量。
func (s *Notification) GetUnreadCount(ctx context.Context, userID uint64) (int64, error) {
	return s.query.GetUnreadCount(ctx, userID)
}

// CreateTemplate 创建一个新的通知模板。
func (s *Notification) CreateTemplate(ctx context.Context, template *domain.NotificationTemplate) error {
	return s.manager.CreateTemplate(ctx, template)
}

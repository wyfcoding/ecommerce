package repository

import (
	"context"

	"ecommerce/internal/notification/model"
)

// NotificationRepo defines the interface for notification data access.
type NotificationRepo interface {
	CreateNotification(ctx context.Context, notification *model.Notification) (*model.Notification, error)
	GetNotificationByID(ctx context.Context, notificationID string) (*model.Notification, error)
	ListNotificationsByUserID(ctx context.Context, userID uint64, includeRead bool, pageSize, pageNum uint32) ([]*model.Notification, uint64, error)
	MarkNotificationAsRead(ctx context.Context, notificationID string) error
}
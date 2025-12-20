package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/notification/domain"
)

// NotificationQuery 处理通知的读操作。
type NotificationQuery struct {
	repo   domain.NotificationRepository
	logger *slog.Logger
}

// NewNotificationQuery 负责处理 NewNotification 相关的读操作和查询逻辑。
func NewNotificationQuery(repo domain.NotificationRepository, logger *slog.Logger) *NotificationQuery {
	return &NotificationQuery{
		repo:   repo,
		logger: logger,
	}
}

// GetNotification 获取指定ID的通知详情。
func (q *NotificationQuery) GetNotification(ctx context.Context, id uint64) (*domain.Notification, error) {
	return q.repo.GetNotification(ctx, id)
}

// ListNotifications 获取用户通知列表。
func (q *NotificationQuery) ListNotifications(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.Notification, int64, error) {
	offset := (page - 1) * pageSize
	var notifStatus *domain.NotificationStatus
	if status != nil {
		s := domain.NotificationStatus(*status)
		notifStatus = &s
	}
	return q.repo.ListNotifications(ctx, userID, notifStatus, offset, pageSize)
}

// GetUnreadCount 获取指定用户的未读通知数量。
func (q *NotificationQuery) GetUnreadCount(ctx context.Context, userID uint64) (int64, error) {
	return q.repo.CountUnreadNotifications(ctx, userID)
}

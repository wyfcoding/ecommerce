package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/notification/domain/entity"
)

// NotificationRepository 通知仓储接口
type NotificationRepository interface {
	// 通知
	SaveNotification(ctx context.Context, notification *entity.Notification) error
	GetNotification(ctx context.Context, id uint64) (*entity.Notification, error)
	ListNotifications(ctx context.Context, userID uint64, status *entity.NotificationStatus, offset, limit int) ([]*entity.Notification, int64, error)
	CountUnreadNotifications(ctx context.Context, userID uint64) (int64, error)

	// 模板
	SaveTemplate(ctx context.Context, template *entity.NotificationTemplate) error
	GetTemplateByCode(ctx context.Context, code string) (*entity.NotificationTemplate, error)
	ListTemplates(ctx context.Context, offset, limit int) ([]*entity.NotificationTemplate, int64, error)
}

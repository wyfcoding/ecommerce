package domain

import (
	"context"
	"time"
)

// NotificationRepository 是通知模块的仓储接口。
// 它定义了对通知和通知模板实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type NotificationRepository interface {
	// 事务管理
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// --- 通知 (Notification methods) ---
	SaveNotification(ctx context.Context, notification *Notification) error
	SaveNotificationInTx(ctx context.Context, tx any, notification *Notification) error
	GetNotification(ctx context.Context, id uint64) (*Notification, error)
	ListNotifications(ctx context.Context, userID uint64, status *NotificationStatus, offset, limit int) ([]*Notification, int64, error)
	CountUnreadNotifications(ctx context.Context, userID uint64) (int64, error)
	DeleteNotification(ctx context.Context, id uint64) error
	DeleteNotificationInTx(ctx context.Context, tx any, id uint64) error
	GetNotificationsByDateRange(ctx context.Context, start, end time.Time) ([]*Notification, error)

	// --- 模板 (NotificationTemplate methods) ---
	SaveTemplate(ctx context.Context, template *NotificationTemplate) error
	SaveTemplateInTx(ctx context.Context, tx any, template *NotificationTemplate) error
	GetTemplateByCode(ctx context.Context, code string) (*NotificationTemplate, error)
	ListTemplates(ctx context.Context, offset, limit int) ([]*NotificationTemplate, int64, error)
}

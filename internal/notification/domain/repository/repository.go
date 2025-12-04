package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/notification/domain/entity" // 导入通知领域的实体定义。
)

// NotificationRepository 是通知模块的仓储接口。
// 它定义了对通知和通知模板实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type NotificationRepository interface {
	// --- 通知 (Notification methods) ---

	// SaveNotification 将通知实体保存到数据存储中。
	// ctx: 上下文。
	// notification: 待保存的通知实体。
	SaveNotification(ctx context.Context, notification *entity.Notification) error
	// GetNotification 根据ID获取通知实体。
	GetNotification(ctx context.Context, id uint64) (*entity.Notification, error)
	// ListNotifications 列出指定用户ID的所有通知实体，支持通过状态过滤和分页。
	ListNotifications(ctx context.Context, userID uint64, status *entity.NotificationStatus, offset, limit int) ([]*entity.Notification, int64, error)
	// CountUnreadNotifications 统计指定用户ID的未读通知数量。
	CountUnreadNotifications(ctx context.Context, userID uint64) (int64, error)

	// --- 模板 (NotificationTemplate methods) ---

	// SaveTemplate 将通知模板实体保存到数据存储中。
	SaveTemplate(ctx context.Context, template *entity.NotificationTemplate) error
	// GetTemplateByCode 根据模板代码获取通知模板实体。
	GetTemplateByCode(ctx context.Context, code string) (*entity.NotificationTemplate, error)
	// ListTemplates 列出所有通知模板实体，支持分页。
	ListTemplates(ctx context.Context, offset, limit int) ([]*entity.NotificationTemplate, int64, error)
}

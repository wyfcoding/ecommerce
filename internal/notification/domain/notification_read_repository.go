// 生成摘要：定义通知读模型仓储接口（Redis），用于高频查询。
package domain

import "context"

// NotificationReadRepository 定义通知读模型接口。
type NotificationReadRepository interface {
	Save(ctx context.Context, notification *Notification) error
	GetByID(ctx context.Context, id uint64) (*Notification, error)
	Delete(ctx context.Context, id uint64) error
}

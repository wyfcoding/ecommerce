// 生成摘要：定义通知搜索仓储接口（Elasticsearch）。
package domain

import "context"

// NotificationSearchRepository 定义通知搜索的访问接口。
type NotificationSearchRepository interface {
	Index(ctx context.Context, notification *Notification) error
	Delete(ctx context.Context, notificationID uint64) error
	Search(ctx context.Context, userID uint64, status *NotificationStatus, offset, limit int) ([]*Notification, int64, error)
}

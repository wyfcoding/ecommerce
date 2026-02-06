// 生成摘要：定义通知模板搜索仓储接口（Elasticsearch）。
package domain

import "context"

// NotificationTemplateSearchRepository 定义通知模板搜索的访问接口。
type NotificationTemplateSearchRepository interface {
	Index(ctx context.Context, template *NotificationTemplate) error
	Delete(ctx context.Context, templateID uint64) error
	Search(ctx context.Context, offset, limit int) ([]*NotificationTemplate, int64, error)
}

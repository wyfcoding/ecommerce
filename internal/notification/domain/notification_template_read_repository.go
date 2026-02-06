// 生成摘要：定义通知模板读模型仓储接口（Redis）。
package domain

import "context"

// NotificationTemplateReadRepository 定义通知模板读模型接口。
type NotificationTemplateReadRepository interface {
	Save(ctx context.Context, template *NotificationTemplate) error
	GetByCode(ctx context.Context, code string) (*NotificationTemplate, error)
	DeleteByCode(ctx context.Context, code string) error
}

// 生成摘要：定义系统配置读模型仓储接口（Redis）。
package domain

import "context"

// SettingReadRepository 读取系统配置缓存接口。
type SettingReadRepository interface {
	Save(ctx context.Context, setting *SystemSetting) error
	GetByKey(ctx context.Context, key string) (*SystemSetting, error)
	Delete(ctx context.Context, key string) error
}

// 生成摘要：定义售后配置读模型仓储接口（Redis）。
package domain

import "context"

// AfterSalesConfigReadRepository 定义售后配置读模型接口。
type AfterSalesConfigReadRepository interface {
	Save(ctx context.Context, cfg *AfterSalesConfig) error
	GetByKey(ctx context.Context, key string) (*AfterSalesConfig, error)
	Delete(ctx context.Context, key string) error
}

// 生成摘要：定义售后读模型仓储接口（Redis），用于高频查询。
package domain

import "context"

// AfterSalesReadRepository 定义售后读模型接口。
type AfterSalesReadRepository interface {
	Save(ctx context.Context, afterSales *AfterSales) error
	GetByID(ctx context.Context, id uint64) (*AfterSales, error)
	GetByNo(ctx context.Context, no string) (*AfterSales, error)
	Delete(ctx context.Context, id uint64, no string) error
}

// 生成摘要：定义售后搜索仓储接口（Elasticsearch）。
package domain

import "context"

// AfterSalesSearchRepository 定义售后搜索访问接口。
type AfterSalesSearchRepository interface {
	Index(ctx context.Context, afterSales *AfterSales) error
	Delete(ctx context.Context, afterSalesID uint64) error
	Search(ctx context.Context, query *AfterSalesQuery) ([]*AfterSales, int64, error)
}

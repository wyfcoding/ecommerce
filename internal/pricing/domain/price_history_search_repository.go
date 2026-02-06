// 生成摘要：定义价格历史搜索仓储接口（Elasticsearch）。
package domain

import "context"

// PriceHistorySearchRepository 定义价格历史搜索接口。
type PriceHistorySearchRepository interface {
	Index(ctx context.Context, history *PriceHistory) error
	Delete(ctx context.Context, id uint64) error
	Search(ctx context.Context, productID, skuID uint64, offset, limit int) ([]*PriceHistory, int64, error)
}

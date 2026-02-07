package domain

import "context"

// SplitOrderSearchRepository 定义拆分订单搜索仓储接口（Elasticsearch）。
type SplitOrderSearchRepository interface {
	Index(ctx context.Context, order *SplitOrder) error
	Delete(ctx context.Context, splitOrderID uint64) error
	SearchByOriginalOrderID(ctx context.Context, originalOrderID uint64, offset, limit int) ([]*SplitOrder, int64, error)
}

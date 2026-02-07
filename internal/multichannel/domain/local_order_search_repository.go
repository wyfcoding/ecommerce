package domain

import "context"

// LocalOrderSearchRepository 定义渠道订单搜索仓储接口（Elasticsearch）。
type LocalOrderSearchRepository interface {
	Index(ctx context.Context, order *LocalOrder) error
	Delete(ctx context.Context, orderID uint64) error
	Search(ctx context.Context, query *LocalOrderQuery, offset, limit int) ([]*LocalOrder, int64, error)
}

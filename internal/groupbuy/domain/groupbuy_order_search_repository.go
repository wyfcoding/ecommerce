package domain

import "context"

// GroupbuyOrderSearchRepository 定义拼团订单搜索仓储接口（Elasticsearch）。
type GroupbuyOrderSearchRepository interface {
	Index(ctx context.Context, order *GroupbuyOrder) error
	Delete(ctx context.Context, orderID uint64) error
	Search(ctx context.Context, query *GroupbuyOrderQuery, offset, limit int) ([]*GroupbuyOrder, int64, error)
}

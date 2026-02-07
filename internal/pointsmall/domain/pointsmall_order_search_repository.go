package domain

import "context"

// PointsOrderSearchRepository 定义积分订单搜索仓储接口（Elasticsearch）。
type PointsOrderSearchRepository interface {
	Index(ctx context.Context, order *PointsOrder) error
	Delete(ctx context.Context, orderID uint64) error
	Search(ctx context.Context, query *PointsOrderQuery, offset, limit int) ([]*PointsOrder, int64, error)
}

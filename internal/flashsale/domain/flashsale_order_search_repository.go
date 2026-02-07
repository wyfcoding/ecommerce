package domain

import "context"

// FlashsaleOrderSearchRepository 定义秒杀订单搜索仓储接口（Elasticsearch）。
type FlashsaleOrderSearchRepository interface {
	Index(ctx context.Context, order *FlashsaleOrder) error
	Delete(ctx context.Context, orderID uint64) error
	Search(ctx context.Context, query *FlashsaleOrderQuery, offset, limit int) ([]*FlashsaleOrder, int64, error)
}

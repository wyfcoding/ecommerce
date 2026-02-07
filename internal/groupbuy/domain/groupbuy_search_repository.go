package domain

import "context"

// GroupbuySearchRepository 定义拼团活动搜索仓储接口（Elasticsearch）。
type GroupbuySearchRepository interface {
	Index(ctx context.Context, groupbuy *Groupbuy) error
	Delete(ctx context.Context, groupbuyID uint64) error
	Search(ctx context.Context, query *GroupbuyQuery, offset, limit int) ([]*Groupbuy, int64, error)
}

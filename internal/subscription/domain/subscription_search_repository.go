package domain

import "context"

// SubscriptionSearchRepository 定义订阅记录搜索仓储接口（Elasticsearch）。
type SubscriptionSearchRepository interface {
	Index(ctx context.Context, sub *Subscription) error
	Delete(ctx context.Context, subscriptionID uint64) error
	Search(ctx context.Context, query *SubscriptionQuery, offset, limit int) ([]*Subscription, int64, error)
}

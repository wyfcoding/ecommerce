package domain

import "context"

// FlashsaleSearchRepository 定义秒杀活动搜索仓储接口（Elasticsearch）。
type FlashsaleSearchRepository interface {
	Index(ctx context.Context, flashsale *Flashsale) error
	Delete(ctx context.Context, flashsaleID uint64) error
	Search(ctx context.Context, query *FlashsaleQuery, offset, limit int) ([]*Flashsale, int64, error)
}

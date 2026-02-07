package domain

import "context"

// SplitOrderReadRepository 定义拆分订单读模型仓储接口（Redis）。
type SplitOrderReadRepository interface {
	Save(ctx context.Context, originalOrderID uint64, orders []*SplitOrder) error
	GetByOriginalOrderID(ctx context.Context, originalOrderID uint64) ([]*SplitOrder, error)
	DeleteByOriginalOrderID(ctx context.Context, originalOrderID uint64) error
}

package domain

import "context"

// DynamicPriceReadRepository 定义动态价格读模型仓储接口（Redis）。
type DynamicPriceReadRepository interface {
	SaveLatest(ctx context.Context, price *DynamicPrice) error
	GetLatest(ctx context.Context, skuID uint64) (*DynamicPrice, error)
	DeleteLatest(ctx context.Context, skuID uint64) error
}

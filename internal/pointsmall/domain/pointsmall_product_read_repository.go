package domain

import "context"

// PointsProductReadRepository 定义积分商品读模型仓储接口（Redis）。
type PointsProductReadRepository interface {
	Save(ctx context.Context, product *PointsProduct) error
	GetByID(ctx context.Context, id uint64) (*PointsProduct, error)
	Delete(ctx context.Context, id uint64) error
}

package domain

import "context"

// PointsOrderReadRepository 定义积分订单读模型仓储接口（Redis）。
type PointsOrderReadRepository interface {
	Save(ctx context.Context, order *PointsOrder) error
	GetByID(ctx context.Context, id uint64) (*PointsOrder, error)
	Delete(ctx context.Context, id uint64) error
}

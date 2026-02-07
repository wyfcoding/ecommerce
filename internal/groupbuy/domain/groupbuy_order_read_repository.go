package domain

import "context"

// GroupbuyOrderReadRepository 定义拼团订单读模型仓储接口（Redis）。
type GroupbuyOrderReadRepository interface {
	Save(ctx context.Context, order *GroupbuyOrder) error
	GetByID(ctx context.Context, id uint64) (*GroupbuyOrder, error)
	Delete(ctx context.Context, id uint64) error
}

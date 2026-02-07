package domain

import "context"

// MergedOrderReadRepository 定义合并订单读模型仓储接口（Redis）。
type MergedOrderReadRepository interface {
	Save(ctx context.Context, order *MergedOrder) error
	GetByID(ctx context.Context, id uint64) (*MergedOrder, error)
	Delete(ctx context.Context, id uint64) error
}

package domain

import "context"

// FlashsaleOrderReadRepository 定义秒杀订单读模型仓储接口（Redis）。
type FlashsaleOrderReadRepository interface {
	Save(ctx context.Context, order *FlashsaleOrder) error
	GetByID(ctx context.Context, id uint64) (*FlashsaleOrder, error)
	Delete(ctx context.Context, id uint64) error
}

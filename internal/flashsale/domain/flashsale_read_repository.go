package domain

import "context"

// FlashsaleReadRepository 定义秒杀活动读模型仓储接口（Redis）。
type FlashsaleReadRepository interface {
	Save(ctx context.Context, flashsale *Flashsale) error
	GetByID(ctx context.Context, id uint64) (*Flashsale, error)
	Delete(ctx context.Context, id uint64) error
}

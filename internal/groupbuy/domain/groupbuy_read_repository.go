package domain

import "context"

// GroupbuyReadRepository 定义拼团活动读模型仓储接口（Redis）。
type GroupbuyReadRepository interface {
	Save(ctx context.Context, groupbuy *Groupbuy) error
	GetByID(ctx context.Context, id uint64) (*Groupbuy, error)
	Delete(ctx context.Context, id uint64) error
}

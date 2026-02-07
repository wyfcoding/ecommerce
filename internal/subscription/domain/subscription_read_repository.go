package domain

import "context"

// SubscriptionReadRepository 定义订阅记录读模型仓储接口（Redis）。
type SubscriptionReadRepository interface {
	Save(ctx context.Context, sub *Subscription) error
	GetByID(ctx context.Context, id uint64) (*Subscription, error)
	Delete(ctx context.Context, id uint64) error
}

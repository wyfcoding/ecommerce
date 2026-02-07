package domain

import "context"

// UserBehaviorReadRepository 定义用户行为读模型仓储接口（Redis）。
type UserBehaviorReadRepository interface {
	Save(ctx context.Context, behavior *UserBehavior) error
	GetByUserID(ctx context.Context, userID uint64) (*UserBehavior, error)
	DeleteByUserID(ctx context.Context, userID uint64) error
}

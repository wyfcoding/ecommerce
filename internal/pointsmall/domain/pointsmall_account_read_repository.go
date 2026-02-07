package domain

import "context"

// PointsAccountReadRepository 定义积分账户读模型仓储接口（Redis）。
type PointsAccountReadRepository interface {
	Save(ctx context.Context, account *PointsAccount) error
	GetByUserID(ctx context.Context, userID uint64) (*PointsAccount, error)
	DeleteByUserID(ctx context.Context, userID uint64) error
}

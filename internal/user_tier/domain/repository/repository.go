package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/user_tier/domain/entity"
)

// UserTierRepository 用户等级仓储接口
type UserTierRepository interface {
	// 用户等级
	SaveUserTier(ctx context.Context, tier *entity.UserTier) error
	GetUserTier(ctx context.Context, userID uint64) (*entity.UserTier, error)

	// 等级配置
	SaveTierConfig(ctx context.Context, config *entity.TierConfig) error
	GetTierConfig(ctx context.Context, level entity.TierLevel) (*entity.TierConfig, error)
	ListTierConfigs(ctx context.Context) ([]*entity.TierConfig, error)

	// 积分
	GetPointsAccount(ctx context.Context, userID uint64) (*entity.PointsAccount, error)
	SavePointsAccount(ctx context.Context, account *entity.PointsAccount) error
	SavePointsLog(ctx context.Context, log *entity.PointsLog) error
	ListPointsLogs(ctx context.Context, userID uint64, offset, limit int) ([]*entity.PointsLog, int64, error)

	// 兑换
	GetExchange(ctx context.Context, id uint64) (*entity.Exchange, error)
	ListExchanges(ctx context.Context, offset, limit int) ([]*entity.Exchange, int64, error)
	SaveExchange(ctx context.Context, exchange *entity.Exchange) error
	SaveExchangeRecord(ctx context.Context, record *entity.ExchangeRecord) error
	ListExchangeRecords(ctx context.Context, userID uint64, offset, limit int) ([]*entity.ExchangeRecord, int64, error)
}

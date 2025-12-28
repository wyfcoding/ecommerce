package domain

import (
	"context"
)

// UserTierRepository 是用户等级模块的仓储接口。
// 它定义了对用户等级、等级配置、积分账户、积分日志、兑换商品和兑换记录实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type UserTierRepository interface {
	// --- 用户等级 (UserTier methods) ---

	// SaveUserTier 将用户等级实体保存到数据存储中。
	// 如果实体已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// tier: 待保存的用户等级实体。
	SaveUserTier(ctx context.Context, tier *UserTier) error
	// GetUserTier 根据用户ID获取用户等级实体。
	GetUserTier(ctx context.Context, userID uint64) (*UserTier, error)

	// --- 等级配置 (TierConfig methods) ---

	// SaveTierConfig 将等级配置实体保存到数据存储中。
	SaveTierConfig(ctx context.Context, config *TierConfig) error
	// GetTierConfig 根据等级级别获取等级配置实体。
	GetTierConfig(ctx context.Context, level TierLevel) (*TierConfig, error)
	// ListTierConfigs 列出所有等级配置实体。
	ListTierConfigs(ctx context.Context) ([]*TierConfig, error)

	// --- 积分 (PointsAccount & PointsLog methods) ---

	// GetPointsAccount 根据用户ID获取积分账户实体。
	GetPointsAccount(ctx context.Context, userID uint64) (*PointsAccount, error)
	// SavePointsAccount 将积分账户实体保存到数据存储中。
	SavePointsAccount(ctx context.Context, account *PointsAccount) error
	// SavePointsLog 将积分日志实体保存到数据存储中。
	SavePointsLog(ctx context.Context, log *PointsLog) error
	// ListPointsLogs 列出指定用户ID的所有积分日志实体，支持分页。
	ListPointsLogs(ctx context.Context, userID uint64, offset, limit int) ([]*PointsLog, int64, error)

	// --- 兑换 (Exchange & ExchangeRecord methods) ---

	// GetExchange 根据ID获取兑换商品实体。
	GetExchange(ctx context.Context, id uint64) (*Exchange, error)
	// ListExchanges 列出所有兑换商品实体，支持分页。
	ListExchanges(ctx context.Context, offset, limit int) ([]*Exchange, int64, error)
	// SaveExchange 将兑换商品实体保存到数据存储中。
	SaveExchange(ctx context.Context, exchange *Exchange) error
	// SaveExchangeRecord 将兑换记录实体保存到数据存储中。
	SaveExchangeRecord(ctx context.Context, record *ExchangeRecord) error
	// ListExchangeRecords 列出指定用户ID的所有兑换记录实体，支持分页。
	ListExchangeRecords(ctx context.Context, userID uint64, offset, limit int) ([]*ExchangeRecord, int64, error)
}

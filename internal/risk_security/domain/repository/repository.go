package repository

import (
	"context"
	"ecommerce/internal/risk_security/domain/entity"
)

// RiskRepository 风险控制仓储接口
type RiskRepository interface {
	// 风险分析记录
	SaveAnalysisResult(ctx context.Context, result *entity.RiskAnalysisResult) error
	GetAnalysisResult(ctx context.Context, id uint64) (*entity.RiskAnalysisResult, error)
	ListAnalysisResults(ctx context.Context, userID uint64, limit int) ([]*entity.RiskAnalysisResult, error)

	// 黑名单
	SaveBlacklist(ctx context.Context, blacklist *entity.Blacklist) error
	GetBlacklist(ctx context.Context, bType entity.BlacklistType, value string) (*entity.Blacklist, error)
	DeleteBlacklist(ctx context.Context, id uint64) error
	IsBlacklisted(ctx context.Context, bType entity.BlacklistType, value string) (bool, error)

	// 设备指纹
	SaveDeviceFingerprint(ctx context.Context, fp *entity.DeviceFingerprint) error
	GetDeviceFingerprint(ctx context.Context, deviceID string) (*entity.DeviceFingerprint, error)

	// 用户行为
	SaveUserBehavior(ctx context.Context, behavior *entity.UserBehavior) error
	GetUserBehavior(ctx context.Context, userID uint64) (*entity.UserBehavior, error)

	// 规则
	ListEnabledRules(ctx context.Context) ([]*entity.RiskRule, error)
}

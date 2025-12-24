package domain

import (
	"context"
)

// RiskRepository 是风控安全模块的仓储接口。
type RiskRepository interface {
	// --- 风险分析记录 (RiskAnalysisResult methods) ---
	SaveAnalysisResult(ctx context.Context, result *RiskAnalysisResult) error
	GetAnalysisResult(ctx context.Context, id uint64) (*RiskAnalysisResult, error)
	ListAnalysisResults(ctx context.Context, userID uint64, limit int) ([]*RiskAnalysisResult, error)

	// --- 黑名单 (Blacklist methods) ---
	SaveBlacklist(ctx context.Context, blacklist *Blacklist) error
	GetBlacklist(ctx context.Context, bType BlacklistType, value string) (*Blacklist, error)
	DeleteBlacklist(ctx context.Context, id uint64) error
	IsBlacklisted(ctx context.Context, bType BlacklistType, value string) (bool, error)

	// --- 设备指纹 (DeviceFingerprint methods) ---
	SaveDeviceFingerprint(ctx context.Context, fp *DeviceFingerprint) error
	GetDeviceFingerprint(ctx context.Context, deviceID string) (*DeviceFingerprint, error)

	// --- 用户行为 (UserBehavior methods) ---
	SaveUserBehavior(ctx context.Context, behavior *UserBehavior) error
	GetUserBehavior(ctx context.Context, userID uint64) (*UserBehavior, error)

	// --- 规则 (RiskRule methods) ---
	ListEnabledRules(ctx context.Context) ([]*RiskRule, error)
}

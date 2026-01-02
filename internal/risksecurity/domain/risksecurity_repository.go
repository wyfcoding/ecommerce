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

	// --- 速度/频次统计 (Velocity Metrics) ---
	GetVelocityMetrics(ctx context.Context, userID uint64) (*VelocityMetrics, error)
}

// VelocityMetrics 用户的交易速度/频次统计指标
type VelocityMetrics struct {
	TxCount1h       int   `json:"tx_count_1h"`
	TxAmount1h      int64 `json:"tx_amount_1h"`
	TxCount24h      int   `json:"tx_count_24h"`
	FailedTxCount1h int   `json:"failed_tx_count_1h"`
}

// FrequencyRepository 定义频率统计的接口
type FrequencyRepository interface {
	// Add 增加计数
	Add(ctx context.Context, key string, delta uint64) error
	// Estimate 获取估计的频率
	Estimate(ctx context.Context, key string) (uint64, error)
	// Reset 重置计数器
	Reset(ctx context.Context) error
}

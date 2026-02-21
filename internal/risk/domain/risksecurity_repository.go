package domain

import (
	"context"
	"time"
)

// RiskRepository 是风控安全模块的写模型仓储接口。
type RiskRepository interface {
	// --- tx helpers ---
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// --- 风险分析记录 (RiskAnalysisResult methods) ---
	SaveAnalysisResult(ctx context.Context, result *RiskAnalysisResult) error
	SaveAnalysisResultInTx(ctx context.Context, tx any, result *RiskAnalysisResult) error
	GetAnalysisResult(ctx context.Context, id uint64) (*RiskAnalysisResult, error)
	ListAnalysisResults(ctx context.Context, query *RiskAnalysisQuery) ([]*RiskAnalysisResult, int64, error)
	DeleteAnalysisResult(ctx context.Context, id uint64) error
	DeleteAnalysisResultInTx(ctx context.Context, tx any, id uint64) error

	// --- 黑名单 (Blacklist methods) ---
	SaveBlacklist(ctx context.Context, blacklist *Blacklist) error
	SaveBlacklistInTx(ctx context.Context, tx any, blacklist *Blacklist) error
	GetBlacklist(ctx context.Context, bType BlacklistType, value string) (*Blacklist, error)
	GetBlacklistByID(ctx context.Context, id uint64) (*Blacklist, error)
	ListBlacklists(ctx context.Context, query *BlacklistQuery) ([]*Blacklist, int64, error)
	DeleteBlacklist(ctx context.Context, id uint64) error
	DeleteBlacklistInTx(ctx context.Context, tx any, id uint64) error
	DeleteBlacklistByTypeAndValue(ctx context.Context, bType BlacklistType, value string) error
	DeleteBlacklistByTypeAndValueInTx(ctx context.Context, tx any, bType BlacklistType, value string) error
	IsBlacklisted(ctx context.Context, bType BlacklistType, value string) (bool, error)

	// --- 设备指纹 (DeviceFingerprint methods) ---
	SaveDeviceFingerprint(ctx context.Context, fp *DeviceFingerprint) error
	SaveDeviceFingerprintInTx(ctx context.Context, tx any, fp *DeviceFingerprint) error
	GetDeviceFingerprint(ctx context.Context, deviceID string) (*DeviceFingerprint, error)
	GetDeviceFingerprintByID(ctx context.Context, id uint64) (*DeviceFingerprint, error)

	// --- 用户行为 (UserBehavior methods) ---
	SaveUserBehavior(ctx context.Context, behavior *UserBehavior) error
	SaveUserBehaviorInTx(ctx context.Context, tx any, behavior *UserBehavior) error
	GetUserBehavior(ctx context.Context, userID uint64) (*UserBehavior, error)

	// --- 规则 (RiskRule methods) ---
	ListEnabledRules(ctx context.Context) ([]*RiskRule, error)
	ListRules(ctx context.Context, enabledOnly bool) ([]*RiskRule, error)
	GetRule(ctx context.Context, id uint64) (*RiskRule, error)
	SaveRule(ctx context.Context, rule *RiskRule) error
	SaveRuleInTx(ctx context.Context, tx any, rule *RiskRule) error
	DeleteRule(ctx context.Context, id uint64) error
	DeleteRuleInTx(ctx context.Context, tx any, id uint64) error

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

// RiskAnalysisQuery 风险分析结果查询条件。
type RiskAnalysisQuery struct {
	UserID    uint64
	Level     *RiskLevel
	StartTime time.Time
	EndTime   time.Time
	Page      int
	PageSize  int
}

// BlacklistQuery 黑名单查询条件。
type BlacklistQuery struct {
	Type       BlacklistType
	ValueLike  string
	ActiveOnly bool
	Page       int
	PageSize   int
}

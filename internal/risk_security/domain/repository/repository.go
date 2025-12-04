package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/risk_security/domain/entity" // 导入风控安全领域的实体定义。
)

// RiskRepository 是风控安全模块的仓储接口。
// 它定义了对风险分析结果、黑名单、设备指纹、用户行为和风险规则实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type RiskRepository interface {
	// --- 风险分析记录 (RiskAnalysisResult methods) ---

	// SaveAnalysisResult 将风险分析结果实体保存到数据存储中。
	// ctx: 上下文。
	// result: 待保存的风险分析结果实体。
	SaveAnalysisResult(ctx context.Context, result *entity.RiskAnalysisResult) error
	// GetAnalysisResult 根据ID获取风险分析结果实体。
	GetAnalysisResult(ctx context.Context, id uint64) (*entity.RiskAnalysisResult, error)
	// ListAnalysisResults 列出指定用户ID的风险分析结果实体，支持数量限制。
	ListAnalysisResults(ctx context.Context, userID uint64, limit int) ([]*entity.RiskAnalysisResult, error)

	// --- 黑名单 (Blacklist methods) ---

	// SaveBlacklist 将黑名单实体保存到数据存储中。
	SaveBlacklist(ctx context.Context, blacklist *entity.Blacklist) error
	// GetBlacklist 根据黑名单类型和值获取黑名单实体。
	GetBlacklist(ctx context.Context, bType entity.BlacklistType, value string) (*entity.Blacklist, error)
	// DeleteBlacklist 根据ID删除黑名单实体。
	DeleteBlacklist(ctx context.Context, id uint64) error
	// IsBlacklisted 检查指定类型和值的实体是否在黑名单中。
	IsBlacklisted(ctx context.Context, bType entity.BlacklistType, value string) (bool, error)

	// --- 设备指纹 (DeviceFingerprint methods) ---

	// SaveDeviceFingerprint 将设备指纹实体保存到数据存储中。
	SaveDeviceFingerprint(ctx context.Context, fp *entity.DeviceFingerprint) error
	// GetDeviceFingerprint 根据设备ID获取设备指纹实体。
	GetDeviceFingerprint(ctx context.Context, deviceID string) (*entity.DeviceFingerprint, error)

	// --- 用户行为 (UserBehavior methods) ---

	// SaveUserBehavior 将用户行为实体保存到数据存储中。
	SaveUserBehavior(ctx context.Context, behavior *entity.UserBehavior) error
	// GetUserBehavior 根据用户ID获取用户行为实体。
	GetUserBehavior(ctx context.Context, userID uint64) (*entity.UserBehavior, error)

	// --- 规则 (RiskRule methods) ---

	// ListEnabledRules 列出所有已启用的风险规则实体。
	ListEnabledRules(ctx context.Context) ([]*entity.RiskRule, error)
}

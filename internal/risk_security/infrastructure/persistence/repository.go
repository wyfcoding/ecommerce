package persistence

import (
	"context"
	"errors" // 导入标准错误处理库。
	"time"   // 导入时间库。

	"github.com/wyfcoding/ecommerce/internal/risk_security/domain/entity"     // 导入风控安全领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/risk_security/domain/repository" // 导入风控安全领域的仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type riskRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewRiskRepository 创建并返回一个新的 riskRepository 实例。
func NewRiskRepository(db *gorm.DB) repository.RiskRepository {
	return &riskRepository{db: db}
}

// --- 风险分析记录 (RiskAnalysisResult methods) ---

// SaveAnalysisResult 将风险分析结果实体保存到数据库。
// 如果实体已存在，则更新；如果不存在，则创建。
func (r *riskRepository) SaveAnalysisResult(ctx context.Context, result *entity.RiskAnalysisResult) error {
	return r.db.WithContext(ctx).Save(result).Error
}

// GetAnalysisResult 根据ID从数据库获取风险分析结果记录。
// 如果记录未找到，则返回nil。
func (r *riskRepository) GetAnalysisResult(ctx context.Context, id uint64) (*entity.RiskAnalysisResult, error) {
	var result entity.RiskAnalysisResult
	if err := r.db.WithContext(ctx).First(&result, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &result, nil
}

// ListAnalysisResults 从数据库列出指定用户ID的风险分析结果记录，支持数量限制。
func (r *riskRepository) ListAnalysisResults(ctx context.Context, userID uint64, limit int) ([]*entity.RiskAnalysisResult, error) {
	var list []*entity.RiskAnalysisResult
	// 按用户ID过滤，按创建时间降序排列，并应用数量限制。
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- 黑名单 (Blacklist methods) ---

// SaveBlacklist 将黑名单实体保存到数据库。
func (r *riskRepository) SaveBlacklist(ctx context.Context, blacklist *entity.Blacklist) error {
	return r.db.WithContext(ctx).Save(blacklist).Error
}

// GetBlacklist 根据黑名单类型和值从数据库获取黑名单记录。
// 如果记录未找到，则返回nil。
func (r *riskRepository) GetBlacklist(ctx context.Context, bType entity.BlacklistType, value string) (*entity.Blacklist, error) {
	var blacklist entity.Blacklist
	// 按类型和值过滤。
	if err := r.db.WithContext(ctx).Where("type = ? AND value = ?", bType, value).First(&blacklist).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &blacklist, nil
}

// DeleteBlacklist 根据ID从数据库删除黑名单记录。
func (r *riskRepository) DeleteBlacklist(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Blacklist{}, id).Error
}

// IsBlacklisted 检查指定类型和值的实体是否在黑名单中且未过期。
func (r *riskRepository) IsBlacklisted(ctx context.Context, bType entity.BlacklistType, value string) (bool, error) {
	var count int64
	now := time.Now()
	// 查询类型和值匹配，并且过期时间晚于当前时间的活跃黑名单条目。
	err := r.db.WithContext(ctx).Model(&entity.Blacklist{}).
		Where("type = ? AND value = ? AND expires_at > ?", bType, value, now).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil // 如果计数大于0，则表示在黑名单中。
}

// --- 设备指纹 (DeviceFingerprint methods) ---

// SaveDeviceFingerprint 将设备指纹实体保存到数据库。
func (r *riskRepository) SaveDeviceFingerprint(ctx context.Context, fp *entity.DeviceFingerprint) error {
	return r.db.WithContext(ctx).Save(fp).Error
}

// GetDeviceFingerprint 根据设备ID从数据库获取设备指纹记录。
// 如果记录未找到，则返回nil。
func (r *riskRepository) GetDeviceFingerprint(ctx context.Context, deviceID string) (*entity.DeviceFingerprint, error) {
	var fp entity.DeviceFingerprint
	// 按设备ID过滤。
	if err := r.db.WithContext(ctx).Where("device_id = ?", deviceID).First(&fp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &fp, nil
}

// --- 用户行为 (UserBehavior methods) ---

// SaveUserBehavior 将用户行为实体保存到数据库。
func (r *riskRepository) SaveUserBehavior(ctx context.Context, behavior *entity.UserBehavior) error {
	return r.db.WithContext(ctx).Save(behavior).Error
}

// GetUserBehavior 根据用户ID从数据库获取用户行为记录。
// 如果记录未找到，则返回nil。
func (r *riskRepository) GetUserBehavior(ctx context.Context, userID uint64) (*entity.UserBehavior, error) {
	var behavior entity.UserBehavior
	// 按用户ID过滤。
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&behavior).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &behavior, nil
}

// --- 规则 (RiskRule methods) ---

// ListEnabledRules 从数据库列出所有已启用的风险规则记录。
func (r *riskRepository) ListEnabledRules(ctx context.Context) ([]*entity.RiskRule, error) {
	var list []*entity.RiskRule
	// 只查询启用的规则。
	if err := r.db.WithContext(ctx).Where("enabled = ?", true).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

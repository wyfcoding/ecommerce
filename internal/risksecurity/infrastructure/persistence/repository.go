package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/risksecurity/domain"

	"gorm.io/gorm"
)

type riskRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewRiskRepository 创建并返回一个新的 riskRepository 实例。
func NewRiskRepository(db *gorm.DB, rdb *redis.Client) domain.RiskRepository {
	return &riskRepository{db: db, redis: rdb}
}

// --- 风险分析记录 (RiskAnalysisResult methods) ---

func (r *riskRepository) SaveAnalysisResult(ctx context.Context, result *domain.RiskAnalysisResult) error {
	return r.db.WithContext(ctx).Save(result).Error
}

func (r *riskRepository) GetAnalysisResult(ctx context.Context, id uint64) (*domain.RiskAnalysisResult, error) {
	var result domain.RiskAnalysisResult
	if err := r.db.WithContext(ctx).First(&result, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

func (r *riskRepository) ListAnalysisResults(ctx context.Context, userID uint64, limit int) ([]*domain.RiskAnalysisResult, error) {
	var list []*domain.RiskAnalysisResult
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- 黑名单 (Blacklist methods) ---

func (r *riskRepository) SaveBlacklist(ctx context.Context, blacklist *domain.Blacklist) error {
	return r.db.WithContext(ctx).Save(blacklist).Error
}

func (r *riskRepository) GetBlacklist(ctx context.Context, bType domain.BlacklistType, value string) (*domain.Blacklist, error) {
	var blacklist domain.Blacklist
	if err := r.db.WithContext(ctx).Where("type = ? AND value = ?", bType, value).First(&blacklist).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &blacklist, nil
}

func (r *riskRepository) DeleteBlacklist(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Blacklist{}, id).Error
}

func (r *riskRepository) IsBlacklisted(ctx context.Context, bType domain.BlacklistType, value string) (bool, error) {
	var count int64
	now := time.Now()
	err := r.db.WithContext(ctx).Model(&domain.Blacklist{}).
		Where("type = ? AND value = ? AND expires_at > ?", bType, value, now).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// --- 设备指纹 (DeviceFingerprint methods) ---

func (r *riskRepository) SaveDeviceFingerprint(ctx context.Context, fp *domain.DeviceFingerprint) error {
	return r.db.WithContext(ctx).Save(fp).Error
}

func (r *riskRepository) GetDeviceFingerprint(ctx context.Context, deviceID string) (*domain.DeviceFingerprint, error) {
	var fp domain.DeviceFingerprint
	if err := r.db.WithContext(ctx).Where("device_id = ?", deviceID).First(&fp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &fp, nil
}

// --- 用户行为 (UserBehavior methods) ---

func (r *riskRepository) SaveUserBehavior(ctx context.Context, behavior *domain.UserBehavior) error {
	return r.db.WithContext(ctx).Save(behavior).Error
}

func (r *riskRepository) GetUserBehavior(ctx context.Context, userID uint64) (*domain.UserBehavior, error) {
	var behavior domain.UserBehavior
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&behavior).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &behavior, nil
}

// --- 规则 (RiskRule methods) ---

func (r *riskRepository) ListEnabledRules(ctx context.Context) ([]*domain.RiskRule, error) {
	var list []*domain.RiskRule
	if err := r.db.WithContext(ctx).Where("enabled = ?", true).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetVelocityMetrics 从 Redis 批量获取用户的交易频率与额度指标。
// 核心逻辑：利用 Pipeline 减少 RTT 损耗，并对不存在的指标项进行零值填充。
func (r *riskRepository) GetVelocityMetrics(ctx context.Context, userID uint64) (*domain.VelocityMetrics, error) {
	// 定义待查询的各项风控频次指标 Key
	keys := []string{
		fmt.Sprintf("risk:velocity:%d:count:1h", userID),
		fmt.Sprintf("risk:velocity:%d:amount:1h", userID),
		fmt.Sprintf("risk:velocity:%d:count:24h", userID),
		fmt.Sprintf("risk:velocity:%d:fail:1h", userID),
	}

	pipe := r.redis.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	// 执行 Pipeline，忽略整体 Error，因为我们会逐个检查 StringCmd 的具体结果
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		// 注意：如果 pipeline 中所有 key 都不存在，Exec 会返回 redis.Nil，这是正常的
	}

	// 解析结果助手函数，处理 redis.Nil 场景
	getInt := func(cmd *redis.StringCmd) int {
		val, err := cmd.Int()
		if err != nil {
			return 0
		}
		return val
	}
	getInt64 := func(cmd *redis.StringCmd) int64 {
		val, err := cmd.Int64()
		if err != nil {
			return 0
		}
		return val
	}

	return &domain.VelocityMetrics{
		TxCount1h:       getInt(cmds[0]),
		TxAmount1h:      getInt64(cmds[1]),
		TxCount24h:      getInt(cmds[2]),
		FailedTxCount1h: getInt(cmds[3]),
	}, nil
}

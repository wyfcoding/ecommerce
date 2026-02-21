package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/risk/domain"
	"gorm.io/gorm"
)

type riskRepository struct {
	db    *gorm.DB
	redis redis.UniversalClient
}

// NewRiskRepository 创建并返回一个新的 riskRepository 实例。
func NewRiskRepository(db *gorm.DB, redisClient redis.UniversalClient) domain.RiskRepository {
	return &riskRepository{db: db, redis: redisClient}
}

// --- tx helpers ---

func (r *riskRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *riskRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *riskRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *riskRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- 风险分析记录 (RiskAnalysisResult methods) ---

func (r *riskRepository) SaveAnalysisResult(ctx context.Context, result *domain.RiskAnalysisResult) error {
	return r.saveAnalysisResultWithTx(ctx, r.db, result)
}

func (r *riskRepository) SaveAnalysisResultInTx(ctx context.Context, tx any, result *domain.RiskAnalysisResult) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveAnalysisResultWithTx(ctx, gormTx, result)
}

func (r *riskRepository) GetAnalysisResult(ctx context.Context, id uint64) (*domain.RiskAnalysisResult, error) {
	var result RiskAnalysisResultModel
	if err := r.db.WithContext(ctx).First(&result, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toRiskAnalysisResult(&result), nil
}

func (r *riskRepository) ListAnalysisResults(ctx context.Context, query *domain.RiskAnalysisQuery) ([]*domain.RiskAnalysisResult, int64, error) {
	var list []*RiskAnalysisResultModel
	var total int64

	db := r.db.WithContext(ctx).Model(&RiskAnalysisResultModel{})
	if query != nil {
		if query.UserID > 0 {
			db = db.Where("user_id = ?", query.UserID)
		}
		if query.Level != nil {
			db = db.Where("risk_level = ?", *query.Level)
		}
		if !query.StartTime.IsZero() {
			db = db.Where("created_at >= ?", query.StartTime)
		}
		if !query.EndTime.IsZero() {
			db = db.Where("created_at <= ?", query.EndTime)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := 1
	pageSize := 20
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.RiskAnalysisResult, len(list))
	for i, item := range list {
		items[i] = toRiskAnalysisResult(item)
	}
	return items, total, nil
}

func (r *riskRepository) DeleteAnalysisResult(ctx context.Context, id uint64) error {
	return r.deleteAnalysisResultWithTx(ctx, r.db, id)
}

func (r *riskRepository) DeleteAnalysisResultInTx(ctx context.Context, tx any, id uint64) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.deleteAnalysisResultWithTx(ctx, gormTx, id)
}

// --- 黑名单 (Blacklist methods) ---

func (r *riskRepository) SaveBlacklist(ctx context.Context, blacklist *domain.Blacklist) error {
	return r.saveBlacklistWithTx(ctx, r.db, blacklist)
}

func (r *riskRepository) SaveBlacklistInTx(ctx context.Context, tx any, blacklist *domain.Blacklist) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveBlacklistWithTx(ctx, gormTx, blacklist)
}

func (r *riskRepository) GetBlacklist(ctx context.Context, bType domain.BlacklistType, value string) (*domain.Blacklist, error) {
	var blacklist BlacklistModel
	if err := r.db.WithContext(ctx).Where("type = ? AND value = ?", bType, value).First(&blacklist).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toBlacklist(&blacklist), nil
}

func (r *riskRepository) GetBlacklistByID(ctx context.Context, id uint64) (*domain.Blacklist, error) {
	var blacklist BlacklistModel
	if err := r.db.WithContext(ctx).First(&blacklist, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toBlacklist(&blacklist), nil
}

func (r *riskRepository) ListBlacklists(ctx context.Context, query *domain.BlacklistQuery) ([]*domain.Blacklist, int64, error) {
	var list []*BlacklistModel
	var total int64

	db := r.db.WithContext(ctx).Model(&BlacklistModel{})
	if query != nil {
		if query.Type != "" {
			db = db.Where("type = ?", query.Type)
		}
		if query.ValueLike != "" {
			db = db.Where("value LIKE ?", "%"+query.ValueLike+"%")
		}
		if query.ActiveOnly {
			db = db.Where("expires_at > ?", time.Now())
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := 1
	pageSize := 20
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.Blacklist, len(list))
	for i, item := range list {
		items[i] = toBlacklist(item)
	}
	return items, total, nil
}

func (r *riskRepository) DeleteBlacklist(ctx context.Context, id uint64) error {
	return r.deleteBlacklistWithTx(ctx, r.db, id)
}

func (r *riskRepository) DeleteBlacklistInTx(ctx context.Context, tx any, id uint64) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.deleteBlacklistWithTx(ctx, gormTx, id)
}

func (r *riskRepository) DeleteBlacklistByTypeAndValue(ctx context.Context, bType domain.BlacklistType, value string) error {
	return r.db.WithContext(ctx).Where("type = ? AND value = ?", bType, value).Delete(&BlacklistModel{}).Error
}

func (r *riskRepository) DeleteBlacklistByTypeAndValueInTx(ctx context.Context, tx any, bType domain.BlacklistType, value string) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.WithContext(ctx).Where("type = ? AND value = ?", bType, value).Delete(&BlacklistModel{}).Error
}

func (r *riskRepository) IsBlacklisted(ctx context.Context, bType domain.BlacklistType, value string) (bool, error) {
	var count int64
	now := time.Now()
	err := r.db.WithContext(ctx).Model(&BlacklistModel{}).
		Where("type = ? AND value = ? AND expires_at > ?", bType, value, now).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// --- 设备指纹 (DeviceFingerprint methods) ---

func (r *riskRepository) SaveDeviceFingerprint(ctx context.Context, fp *domain.DeviceFingerprint) error {
	return r.saveDeviceFingerprintWithTx(ctx, r.db, fp)
}

func (r *riskRepository) SaveDeviceFingerprintInTx(ctx context.Context, tx any, fp *domain.DeviceFingerprint) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveDeviceFingerprintWithTx(ctx, gormTx, fp)
}

func (r *riskRepository) GetDeviceFingerprint(ctx context.Context, deviceID string) (*domain.DeviceFingerprint, error) {
	var fp DeviceFingerprintModel
	if err := r.db.WithContext(ctx).Where("device_id = ?", deviceID).First(&fp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDeviceFingerprint(&fp), nil
}

func (r *riskRepository) GetDeviceFingerprintByID(ctx context.Context, id uint64) (*domain.DeviceFingerprint, error) {
	var fp DeviceFingerprintModel
	if err := r.db.WithContext(ctx).First(&fp, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDeviceFingerprint(&fp), nil
}

// --- 用户行为 (UserBehavior methods) ---

func (r *riskRepository) SaveUserBehavior(ctx context.Context, behavior *domain.UserBehavior) error {
	return r.saveUserBehaviorWithTx(ctx, r.db, behavior)
}

func (r *riskRepository) SaveUserBehaviorInTx(ctx context.Context, tx any, behavior *domain.UserBehavior) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveUserBehaviorWithTx(ctx, gormTx, behavior)
}

func (r *riskRepository) GetUserBehavior(ctx context.Context, userID uint64) (*domain.UserBehavior, error) {
	var behavior UserBehaviorModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&behavior).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toUserBehavior(&behavior), nil
}

// --- 规则 (RiskRule methods) ---

func (r *riskRepository) ListEnabledRules(ctx context.Context) ([]*domain.RiskRule, error) {
	var list []*RiskRuleModel
	if err := r.db.WithContext(ctx).Where("enabled = ?", true).Find(&list).Error; err != nil {
		return nil, err
	}
	items := make([]*domain.RiskRule, len(list))
	for i, item := range list {
		items[i] = toRiskRule(item)
	}
	return items, nil
}

func (r *riskRepository) ListRules(ctx context.Context, enabledOnly bool) ([]*domain.RiskRule, error) {
	var list []*RiskRuleModel
	db := r.db.WithContext(ctx).Model(&RiskRuleModel{})
	if enabledOnly {
		db = db.Where("enabled = ?", true)
	}
	if err := db.Order("created_at desc").Find(&list).Error; err != nil {
		return nil, err
	}
	items := make([]*domain.RiskRule, len(list))
	for i, item := range list {
		items[i] = toRiskRule(item)
	}
	return items, nil
}

func (r *riskRepository) GetRule(ctx context.Context, id uint64) (*domain.RiskRule, error) {
	var rule RiskRuleModel
	if err := r.db.WithContext(ctx).First(&rule, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toRiskRule(&rule), nil
}

func (r *riskRepository) SaveRule(ctx context.Context, rule *domain.RiskRule) error {
	return r.saveRuleWithTx(ctx, r.db, rule)
}

func (r *riskRepository) SaveRuleInTx(ctx context.Context, tx any, rule *domain.RiskRule) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveRuleWithTx(ctx, gormTx, rule)
}

func (r *riskRepository) DeleteRule(ctx context.Context, id uint64) error {
	return r.deleteRuleWithTx(ctx, r.db, id)
}

func (r *riskRepository) DeleteRuleInTx(ctx context.Context, tx any, id uint64) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.deleteRuleWithTx(ctx, gormTx, id)
}

// --- 速度/频次统计 (Velocity Metrics) ---

func (r *riskRepository) GetVelocityMetrics(ctx context.Context, userID uint64) (*domain.VelocityMetrics, error) {
	if r.redis == nil {
		return &domain.VelocityMetrics{}, nil
	}
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

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return &domain.VelocityMetrics{}, err
	}

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

// --- internal helpers ---

func (r *riskRepository) saveAnalysisResultWithTx(ctx context.Context, tx *gorm.DB, result *domain.RiskAnalysisResult) error {
	if result == nil {
		return nil
	}
	model := toRiskAnalysisResultModel(result)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toRiskAnalysisResult(model); synced != nil {
		*result = *synced
	}
	return nil
}

func (r *riskRepository) deleteAnalysisResultWithTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	return tx.WithContext(ctx).Delete(&RiskAnalysisResultModel{}, id).Error
}

func (r *riskRepository) saveBlacklistWithTx(ctx context.Context, tx *gorm.DB, blacklist *domain.Blacklist) error {
	if blacklist == nil {
		return nil
	}
	model := toBlacklistModel(blacklist)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toBlacklist(model); synced != nil {
		*blacklist = *synced
	}
	return nil
}

func (r *riskRepository) deleteBlacklistWithTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	return tx.WithContext(ctx).Delete(&BlacklistModel{}, id).Error
}

func (r *riskRepository) saveDeviceFingerprintWithTx(ctx context.Context, tx *gorm.DB, fp *domain.DeviceFingerprint) error {
	if fp == nil {
		return nil
	}
	model := toDeviceFingerprintModel(fp)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toDeviceFingerprint(model); synced != nil {
		*fp = *synced
	}
	return nil
}

func (r *riskRepository) saveUserBehaviorWithTx(ctx context.Context, tx *gorm.DB, behavior *domain.UserBehavior) error {
	if behavior == nil {
		return nil
	}
	model := toUserBehaviorModel(behavior)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toUserBehavior(model); synced != nil {
		*behavior = *synced
	}
	return nil
}

func (r *riskRepository) saveRuleWithTx(ctx context.Context, tx *gorm.DB, rule *domain.RiskRule) error {
	if rule == nil {
		return nil
	}
	model := toRiskRuleModel(rule)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toRiskRule(model); synced != nil {
		*rule = *synced
	}
	return nil
}

func (r *riskRepository) deleteRuleWithTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	return tx.WithContext(ctx).Delete(&RiskRuleModel{}, id).Error
}

package persistence

import (
	"context"
	"ecommerce/internal/risk_security/domain/entity"
	"ecommerce/internal/risk_security/domain/repository"
	"errors"
	"time"

	"gorm.io/gorm"
)

type riskRepository struct {
	db *gorm.DB
}

func NewRiskRepository(db *gorm.DB) repository.RiskRepository {
	return &riskRepository{db: db}
}

// 风险分析记录
func (r *riskRepository) SaveAnalysisResult(ctx context.Context, result *entity.RiskAnalysisResult) error {
	return r.db.WithContext(ctx).Save(result).Error
}

func (r *riskRepository) GetAnalysisResult(ctx context.Context, id uint64) (*entity.RiskAnalysisResult, error) {
	var result entity.RiskAnalysisResult
	if err := r.db.WithContext(ctx).First(&result, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

func (r *riskRepository) ListAnalysisResults(ctx context.Context, userID uint64, limit int) ([]*entity.RiskAnalysisResult, error) {
	var list []*entity.RiskAnalysisResult
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// 黑名单
func (r *riskRepository) SaveBlacklist(ctx context.Context, blacklist *entity.Blacklist) error {
	return r.db.WithContext(ctx).Save(blacklist).Error
}

func (r *riskRepository) GetBlacklist(ctx context.Context, bType entity.BlacklistType, value string) (*entity.Blacklist, error) {
	var blacklist entity.Blacklist
	if err := r.db.WithContext(ctx).Where("type = ? AND value = ?", bType, value).First(&blacklist).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &blacklist, nil
}

func (r *riskRepository) DeleteBlacklist(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Blacklist{}, id).Error
}

func (r *riskRepository) IsBlacklisted(ctx context.Context, bType entity.BlacklistType, value string) (bool, error) {
	var count int64
	now := time.Now()
	err := r.db.WithContext(ctx).Model(&entity.Blacklist{}).
		Where("type = ? AND value = ? AND expires_at > ?", bType, value, now).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 设备指纹
func (r *riskRepository) SaveDeviceFingerprint(ctx context.Context, fp *entity.DeviceFingerprint) error {
	return r.db.WithContext(ctx).Save(fp).Error
}

func (r *riskRepository) GetDeviceFingerprint(ctx context.Context, deviceID string) (*entity.DeviceFingerprint, error) {
	var fp entity.DeviceFingerprint
	if err := r.db.WithContext(ctx).Where("device_id = ?", deviceID).First(&fp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &fp, nil
}

// 用户行为
func (r *riskRepository) SaveUserBehavior(ctx context.Context, behavior *entity.UserBehavior) error {
	return r.db.WithContext(ctx).Save(behavior).Error
}

func (r *riskRepository) GetUserBehavior(ctx context.Context, userID uint64) (*entity.UserBehavior, error) {
	var behavior entity.UserBehavior
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&behavior).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &behavior, nil
}

// 规则
func (r *riskRepository) ListEnabledRules(ctx context.Context) ([]*entity.RiskRule, error) {
	var list []*entity.RiskRule
	if err := r.db.WithContext(ctx).Where("enabled = ?", true).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

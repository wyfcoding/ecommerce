package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/usertier/domain"
	"gorm.io/gorm"
)

type userTierRepository struct {
	db *gorm.DB
}

// NewUserTierRepository 创建并返回一个新的 userTierRepository 实例。
func NewUserTierRepository(db *gorm.DB) domain.UserTierRepository {
	return &userTierRepository{db: db}
}

// --- User Tier methods ---

func (r *userTierRepository) SaveUserTier(ctx context.Context, tier *domain.UserTier) error {
	return r.db.WithContext(ctx).Save(tier).Error
}

func (r *userTierRepository) GetUserTier(ctx context.Context, userID uint64) (*domain.UserTier, error) {
	var tier domain.UserTier
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&tier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tier, nil
}

// --- Tier Config methods ---

func (r *userTierRepository) SaveTierConfig(ctx context.Context, config *domain.TierConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

func (r *userTierRepository) GetTierConfig(ctx context.Context, level domain.TierLevel) (*domain.TierConfig, error) {
	var config domain.TierConfig
	if err := r.db.WithContext(ctx).Where("level = ?", level).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

func (r *userTierRepository) ListTierConfigs(ctx context.Context) ([]*domain.TierConfig, error) {
	var list []*domain.TierConfig
	if err := r.db.WithContext(ctx).Order("level asc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- Points methods ---

func (r *userTierRepository) GetPointsAccount(ctx context.Context, userID uint64) (*domain.PointsAccount, error) {
	var account domain.PointsAccount
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *userTierRepository) SavePointsAccount(ctx context.Context, account *domain.PointsAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *userTierRepository) SavePointsLog(ctx context.Context, log *domain.PointsLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

func (r *userTierRepository) ListPointsLogs(ctx context.Context, userID uint64, offset, limit int) ([]*domain.PointsLog, int64, error) {
	var list []*domain.PointsLog
	var total int64
	db := r.db.WithContext(ctx).Model(&domain.PointsLog{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// --- Exchange methods ---

func (r *userTierRepository) GetExchange(ctx context.Context, id uint64) (*domain.Exchange, error) {
	var item domain.Exchange
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *userTierRepository) ListExchanges(ctx context.Context, offset, limit int) ([]*domain.Exchange, int64, error) {
	var list []*domain.Exchange
	var total int64
	db := r.db.WithContext(ctx).Model(&domain.Exchange{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *userTierRepository) SaveExchange(ctx context.Context, exchange *domain.Exchange) error {
	return r.db.WithContext(ctx).Save(exchange).Error
}

func (r *userTierRepository) SaveExchangeRecord(ctx context.Context, record *domain.ExchangeRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

func (r *userTierRepository) ListExchangeRecords(ctx context.Context, userID uint64, offset, limit int) ([]*domain.ExchangeRecord, int64, error) {
	var list []*domain.ExchangeRecord
	var total int64
	db := r.db.WithContext(ctx).Model(&domain.ExchangeRecord{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

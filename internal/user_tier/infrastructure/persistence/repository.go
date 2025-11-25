package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/user_tier/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/user_tier/domain/repository"
	"errors"

	"gorm.io/gorm"
)

type userTierRepository struct {
	db *gorm.DB
}

func NewUserTierRepository(db *gorm.DB) repository.UserTierRepository {
	return &userTierRepository{db: db}
}

// 用户等级
func (r *userTierRepository) SaveUserTier(ctx context.Context, tier *entity.UserTier) error {
	// Upsert based on UserID
	var existing entity.UserTier
	err := r.db.WithContext(ctx).Where("user_id = ?", tier.UserID).First(&existing).Error
	if err == nil {
		tier.ID = existing.ID
		tier.CreatedAt = existing.CreatedAt
		return r.db.WithContext(ctx).Save(tier).Error
	}
	return r.db.WithContext(ctx).Create(tier).Error
}

func (r *userTierRepository) GetUserTier(ctx context.Context, userID uint64) (*entity.UserTier, error) {
	var tier entity.UserTier
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&tier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tier, nil
}

// 等级配置
func (r *userTierRepository) SaveTierConfig(ctx context.Context, config *entity.TierConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

func (r *userTierRepository) GetTierConfig(ctx context.Context, level entity.TierLevel) (*entity.TierConfig, error) {
	var config entity.TierConfig
	if err := r.db.WithContext(ctx).Where("level = ?", level).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

func (r *userTierRepository) ListTierConfigs(ctx context.Context) ([]*entity.TierConfig, error) {
	var list []*entity.TierConfig
	if err := r.db.WithContext(ctx).Order("level asc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// 积分
func (r *userTierRepository) GetPointsAccount(ctx context.Context, userID uint64) (*entity.PointsAccount, error) {
	var account entity.PointsAccount
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *userTierRepository) SavePointsAccount(ctx context.Context, account *entity.PointsAccount) error {
	var existing entity.PointsAccount
	err := r.db.WithContext(ctx).Where("user_id = ?", account.UserID).First(&existing).Error
	if err == nil {
		account.ID = existing.ID
		account.CreatedAt = existing.CreatedAt
		return r.db.WithContext(ctx).Save(account).Error
	}
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *userTierRepository) SavePointsLog(ctx context.Context, log *entity.PointsLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *userTierRepository) ListPointsLogs(ctx context.Context, userID uint64, offset, limit int) ([]*entity.PointsLog, int64, error) {
	var list []*entity.PointsLog
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.PointsLog{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// 兑换
func (r *userTierRepository) GetExchange(ctx context.Context, id uint64) (*entity.Exchange, error) {
	var exchange entity.Exchange
	if err := r.db.WithContext(ctx).First(&exchange, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &exchange, nil
}

func (r *userTierRepository) ListExchanges(ctx context.Context, offset, limit int) ([]*entity.Exchange, int64, error) {
	var list []*entity.Exchange
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Exchange{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *userTierRepository) SaveExchange(ctx context.Context, exchange *entity.Exchange) error {
	return r.db.WithContext(ctx).Save(exchange).Error
}

func (r *userTierRepository) SaveExchangeRecord(ctx context.Context, record *entity.ExchangeRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *userTierRepository) ListExchangeRecords(ctx context.Context, userID uint64, offset, limit int) ([]*entity.ExchangeRecord, int64, error) {
	var list []*entity.ExchangeRecord
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.ExchangeRecord{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

package persistence

import (
	"context"
	"ecommerce/internal/pricing/domain/entity"
	"ecommerce/internal/pricing/domain/repository"
	"errors"
	"time"

	"gorm.io/gorm"
)

type pricingRepository struct {
	db *gorm.DB
}

func NewPricingRepository(db *gorm.DB) repository.PricingRepository {
	return &pricingRepository{db: db}
}

// 规则
func (r *pricingRepository) SaveRule(ctx context.Context, rule *entity.PricingRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

func (r *pricingRepository) GetRule(ctx context.Context, id uint64) (*entity.PricingRule, error) {
	var rule entity.PricingRule
	if err := r.db.WithContext(ctx).First(&rule, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rule, nil
}

func (r *pricingRepository) GetActiveRule(ctx context.Context, productID, skuID uint64) (*entity.PricingRule, error) {
	var rule entity.PricingRule
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND sku_id = ? AND enabled = ? AND start_time <= ? AND end_time >= ?",
			productID, skuID, true, now, now).
		Order("updated_at desc").
		First(&rule).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rule, nil
}

func (r *pricingRepository) ListRules(ctx context.Context, productID uint64, offset, limit int) ([]*entity.PricingRule, int64, error) {
	var list []*entity.PricingRule
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.PricingRule{})
	if productID > 0 {
		db = db.Where("product_id = ?", productID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// 历史
func (r *pricingRepository) SaveHistory(ctx context.Context, history *entity.PriceHistory) error {
	return r.db.WithContext(ctx).Save(history).Error
}

func (r *pricingRepository) ListHistory(ctx context.Context, productID, skuID uint64, offset, limit int) ([]*entity.PriceHistory, int64, error) {
	var list []*entity.PriceHistory
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.PriceHistory{})
	if productID > 0 {
		db = db.Where("product_id = ?", productID)
	}
	if skuID > 0 {
		db = db.Where("sku_id = ?", skuID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/pricing/domain"
	"gorm.io/gorm"
)

type pricingRepository struct {
	db *gorm.DB
}

// NewPricingRepository 创建并返回一个新的 pricingRepository 实例。
func NewPricingRepository(db *gorm.DB) domain.PricingRepository {
	return &pricingRepository{db: db}
}

// --- 规则 (PricingRule methods) ---

func (r *pricingRepository) SaveRule(ctx context.Context, rule *domain.PricingRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

func (r *pricingRepository) GetRule(ctx context.Context, id uint64) (*domain.PricingRule, error) {
	var rule domain.PricingRule
	if err := r.db.WithContext(ctx).First(&rule, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err
	}
	return &rule, nil
}

func (r *pricingRepository) GetActiveRule(ctx context.Context, productID, skuID uint64) (*domain.PricingRule, error) {
	var rule domain.PricingRule
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

func (r *pricingRepository) ListRules(ctx context.Context, productID uint64, offset, limit int) ([]*domain.PricingRule, int64, error) {
	var list []*domain.PricingRule
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.PricingRule{})
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

// --- 历史 (PriceHistory methods) ---

func (r *pricingRepository) SaveHistory(ctx context.Context, history *domain.PriceHistory) error {
	return r.db.WithContext(ctx).Save(history).Error
}

func (r *pricingRepository) ListHistory(ctx context.Context, productID, skuID uint64, offset, limit int) ([]*domain.PriceHistory, int64, error) {
	var list []*domain.PriceHistory
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.PriceHistory{})
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

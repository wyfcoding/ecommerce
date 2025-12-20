package persistence

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/domain"

	"gorm.io/gorm"
)

type pricingRepository struct {
	db *gorm.DB
}

// NewPricingRepository 创建并返回一个新的 pricingRepository 实例。
func NewPricingRepository(db *gorm.DB) domain.PricingRepository {
	return &pricingRepository{db: db}
}

// --- DynamicPrice methods ---

func (r *pricingRepository) SaveDynamicPrice(ctx context.Context, price *domain.DynamicPrice) error {
	return r.db.WithContext(ctx).Save(price).Error
}

func (r *pricingRepository) GetLatestDynamicPrice(ctx context.Context, skuID uint64) (*domain.DynamicPrice, error) {
	var price domain.DynamicPrice
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).Order("created_at desc").First(&price).Error; err != nil {
		return nil, err
	}
	return &price, nil
}

func (r *pricingRepository) ListDynamicPrices(ctx context.Context, skuID uint64, offset, limit int) ([]*domain.DynamicPrice, int64, error) {
	var list []*domain.DynamicPrice
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.DynamicPrice{})
	if skuID != 0 {
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

// --- CompetitorPrice methods ---

func (r *pricingRepository) SaveCompetitorPriceInfo(ctx context.Context, info *domain.CompetitorPriceInfo) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(info).Error; err != nil {
			return err
		}
		for _, comp := range info.Competitors {
			comp.InfoID = uint64(info.ID)
			if err := tx.Save(comp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *pricingRepository) GetCompetitorPriceInfo(ctx context.Context, skuID uint64) (*domain.CompetitorPriceInfo, error) {
	var info domain.CompetitorPriceInfo
	if err := r.db.WithContext(ctx).Preload("Competitors").Where("sku_id = ?", skuID).Order("created_at desc").First(&info).Error; err != nil {
		return nil, err
	}
	return &info, nil
}

// --- PriceHistory methods ---

func (r *pricingRepository) SavePriceHistory(ctx context.Context, history *domain.PriceHistoryData) error {
	return r.db.WithContext(ctx).Save(history).Error
}

func (r *pricingRepository) GetPriceHistory(ctx context.Context, skuID uint64, days int) ([]*domain.PriceHistoryData, error) {
	var list []*domain.PriceHistoryData
	startTime := time.Now().AddDate(0, 0, -days)
	if err := r.db.WithContext(ctx).Where("sku_id = ? AND date >= ?", skuID, startTime).Order("date asc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- PricingStrategy methods ---

func (r *pricingRepository) SavePricingStrategy(ctx context.Context, strategy *domain.PricingStrategy) error {
	return r.db.WithContext(ctx).Save(strategy).Error
}

func (r *pricingRepository) GetPricingStrategy(ctx context.Context, skuID uint64) (*domain.PricingStrategy, error) {
	var strategy domain.PricingStrategy
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).First(&strategy).Error; err != nil {
		return nil, err
	}
	return &strategy, nil
}

func (r *pricingRepository) ListPricingStrategies(ctx context.Context, offset, limit int) ([]*domain.PricingStrategy, int64, error) {
	var list []*domain.PricingStrategy
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.PricingStrategy{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

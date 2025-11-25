package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/domain/repository"
	"time"

	"gorm.io/gorm"
)

type pricingRepository struct {
	db *gorm.DB
}

func NewPricingRepository(db *gorm.DB) repository.PricingRepository {
	return &pricingRepository{db: db}
}

// DynamicPrice methods
func (r *pricingRepository) SaveDynamicPrice(ctx context.Context, price *entity.DynamicPrice) error {
	return r.db.WithContext(ctx).Save(price).Error
}

func (r *pricingRepository) GetLatestDynamicPrice(ctx context.Context, skuID uint64) (*entity.DynamicPrice, error) {
	var price entity.DynamicPrice
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).Order("created_at desc").First(&price).Error; err != nil {
		return nil, err
	}
	return &price, nil
}

func (r *pricingRepository) ListDynamicPrices(ctx context.Context, skuID uint64, offset, limit int) ([]*entity.DynamicPrice, int64, error) {
	var list []*entity.DynamicPrice
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.DynamicPrice{})
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

// CompetitorPrice methods
func (r *pricingRepository) SaveCompetitorPriceInfo(ctx context.Context, info *entity.CompetitorPriceInfo) error {
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

func (r *pricingRepository) GetCompetitorPriceInfo(ctx context.Context, skuID uint64) (*entity.CompetitorPriceInfo, error) {
	var info entity.CompetitorPriceInfo
	if err := r.db.WithContext(ctx).Preload("Competitors").Where("sku_id = ?", skuID).Order("created_at desc").First(&info).Error; err != nil {
		return nil, err
	}
	return &info, nil
}

// PriceHistory methods
func (r *pricingRepository) SavePriceHistory(ctx context.Context, history *entity.PriceHistoryData) error {
	return r.db.WithContext(ctx).Save(history).Error
}

func (r *pricingRepository) GetPriceHistory(ctx context.Context, skuID uint64, days int) ([]*entity.PriceHistoryData, error) {
	var list []*entity.PriceHistoryData
	startTime := time.Now().AddDate(0, 0, -days)
	if err := r.db.WithContext(ctx).Where("sku_id = ? AND date >= ?", skuID, startTime).Order("date asc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// PricingStrategy methods
func (r *pricingRepository) SavePricingStrategy(ctx context.Context, strategy *entity.PricingStrategy) error {
	return r.db.WithContext(ctx).Save(strategy).Error
}

func (r *pricingRepository) GetPricingStrategy(ctx context.Context, skuID uint64) (*entity.PricingStrategy, error) {
	var strategy entity.PricingStrategy
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).First(&strategy).Error; err != nil {
		return nil, err
	}
	return &strategy, nil
}

func (r *pricingRepository) ListPricingStrategies(ctx context.Context, offset, limit int) ([]*entity.PricingStrategy, int64, error) {
	var list []*entity.PricingStrategy
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.PricingStrategy{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

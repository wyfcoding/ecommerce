package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/dynamicpricing/domain"

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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &price, nil
}

// --- CompetitorPrice methods ---

func (r *pricingRepository) GetCompetitorPriceInfo(ctx context.Context, skuID uint64) (*domain.CompetitorPriceInfo, error) {
	var info domain.CompetitorPriceInfo
	if err := r.db.WithContext(ctx).Preload("Competitors").Where("sku_id = ?", skuID).Order("created_at desc").First(&info).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &info, nil
}

// --- PriceHistory methods ---

func (r *pricingRepository) GetPriceHistory(ctx context.Context, skuID uint64, limit int) ([]domain.PriceHistoryData, error) {
	var list []domain.PriceHistoryData
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).Order("date desc").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- PriceElasticity methods ---

func (r *pricingRepository) GetPriceElasticity(ctx context.Context, skuID uint64) (*domain.PriceElasticity, error) {
	var elasticity domain.PriceElasticity
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).Order("created_at desc").First(&elasticity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &elasticity, nil
}

// --- PricingStrategy methods ---

func (r *pricingRepository) SavePricingStrategy(ctx context.Context, strategy *domain.PricingStrategy) error {
	return r.db.WithContext(ctx).Save(strategy).Error
}

func (r *pricingRepository) GetPricingStrategy(ctx context.Context, skuID uint64) (*domain.PricingStrategy, error) {
	var strategy domain.PricingStrategy
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).First(&strategy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
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

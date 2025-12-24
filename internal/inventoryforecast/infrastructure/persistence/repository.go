package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/inventoryforecast/domain"

	"gorm.io/gorm"
)

type inventoryForecastRepository struct {
	db *gorm.DB
}

// NewInventoryForecastRepository 创建并返回一个新的 inventoryForecastRepository 实例。
func NewInventoryForecastRepository(db *gorm.DB) domain.InventoryForecastRepository {
	return &inventoryForecastRepository{db: db}
}

// --- 销售预测 ---

func (r *inventoryForecastRepository) SaveForecast(ctx context.Context, forecast *domain.SalesForecast) error {
	return r.db.WithContext(ctx).Save(forecast).Error
}

func (r *inventoryForecastRepository) GetForecastBySKU(ctx context.Context, skuID uint64) (*domain.SalesForecast, error) {
	var forecast domain.SalesForecast
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).First(&forecast).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &forecast, nil
}

// --- 库存预警 ---

func (r *inventoryForecastRepository) SaveWarning(ctx context.Context, warning *domain.InventoryWarning) error {
	return r.db.WithContext(ctx).Save(warning).Error
}

func (r *inventoryForecastRepository) ListWarnings(ctx context.Context, offset, limit int) ([]*domain.InventoryWarning, int64, error) {
	var list []*domain.InventoryWarning
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.InventoryWarning{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 滞销品 ---

func (r *inventoryForecastRepository) SaveSlowMovingItem(ctx context.Context, item *domain.SlowMovingItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *inventoryForecastRepository) ListSlowMovingItems(ctx context.Context, offset, limit int) ([]*domain.SlowMovingItem, int64, error) {
	var list []*domain.SlowMovingItem
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.SlowMovingItem{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("days_in_stock desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 补货建议 ---

func (r *inventoryForecastRepository) SaveReplenishmentSuggestion(ctx context.Context, suggestion *domain.ReplenishmentSuggestion) error {
	return r.db.WithContext(ctx).Save(suggestion).Error
}

func (r *inventoryForecastRepository) ListReplenishmentSuggestions(ctx context.Context, priority string, offset, limit int) ([]*domain.ReplenishmentSuggestion, int64, error) {
	var list []*domain.ReplenishmentSuggestion
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.ReplenishmentSuggestion{})
	if priority != "" {
		db = db.Where("priority = ?", priority)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 缺货风险 ---

func (r *inventoryForecastRepository) SaveStockoutRisk(ctx context.Context, risk *domain.StockoutRisk) error {
	return r.db.WithContext(ctx).Save(risk).Error
}

func (r *inventoryForecastRepository) ListStockoutRisks(ctx context.Context, level domain.StockoutRiskLevel, offset, limit int) ([]*domain.StockoutRisk, int64, error) {
	var list []*domain.StockoutRisk
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.StockoutRisk{})
	if level != "" {
		db = db.Where("risk_level = ?", level)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("days_until_stockout asc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

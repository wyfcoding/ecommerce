package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/domain/repository"
	"errors"

	"gorm.io/gorm"
)

type inventoryForecastRepository struct {
	db *gorm.DB
}

func NewInventoryForecastRepository(db *gorm.DB) repository.InventoryForecastRepository {
	return &inventoryForecastRepository{db: db}
}

// 销售预测
func (r *inventoryForecastRepository) SaveForecast(ctx context.Context, forecast *entity.SalesForecast) error {
	return r.db.WithContext(ctx).Save(forecast).Error
}

func (r *inventoryForecastRepository) GetForecastBySKU(ctx context.Context, skuID uint64) (*entity.SalesForecast, error) {
	var forecast entity.SalesForecast
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).First(&forecast).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &forecast, nil
}

// 库存预警
func (r *inventoryForecastRepository) SaveWarning(ctx context.Context, warning *entity.InventoryWarning) error {
	return r.db.WithContext(ctx).Save(warning).Error
}

func (r *inventoryForecastRepository) ListWarnings(ctx context.Context, offset, limit int) ([]*entity.InventoryWarning, int64, error) {
	var list []*entity.InventoryWarning
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.InventoryWarning{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// 滞销品
func (r *inventoryForecastRepository) SaveSlowMovingItem(ctx context.Context, item *entity.SlowMovingItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *inventoryForecastRepository) ListSlowMovingItems(ctx context.Context, offset, limit int) ([]*entity.SlowMovingItem, int64, error) {
	var list []*entity.SlowMovingItem
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.SlowMovingItem{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("days_in_stock desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// 补货建议
func (r *inventoryForecastRepository) SaveReplenishmentSuggestion(ctx context.Context, suggestion *entity.ReplenishmentSuggestion) error {
	return r.db.WithContext(ctx).Save(suggestion).Error
}

func (r *inventoryForecastRepository) ListReplenishmentSuggestions(ctx context.Context, priority string, offset, limit int) ([]*entity.ReplenishmentSuggestion, int64, error) {
	var list []*entity.ReplenishmentSuggestion
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.ReplenishmentSuggestion{})
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

// 缺货风险
func (r *inventoryForecastRepository) SaveStockoutRisk(ctx context.Context, risk *entity.StockoutRisk) error {
	return r.db.WithContext(ctx).Save(risk).Error
}

func (r *inventoryForecastRepository) ListStockoutRisks(ctx context.Context, level entity.StockoutRiskLevel, offset, limit int) ([]*entity.StockoutRisk, int64, error) {
	var list []*entity.StockoutRisk
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.StockoutRisk{})
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

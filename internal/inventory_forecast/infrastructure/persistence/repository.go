package persistence

import (
	"context"
	"errors"
	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/domain/entity"     // 导入库存预测模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/domain/repository" // 导入库存预测模块的领域仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type inventoryForecastRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewInventoryForecastRepository 创建并返回一个新的 inventoryForecastRepository 实例。
// db: GORM数据库连接实例。
func NewInventoryForecastRepository(db *gorm.DB) repository.InventoryForecastRepository {
	return &inventoryForecastRepository{db: db}
}

// --- 销售预测 (SalesForecast methods) ---

// SaveForecast 将销售预测实体保存到数据库。
// 如果预测已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *inventoryForecastRepository) SaveForecast(ctx context.Context, forecast *entity.SalesForecast) error {
	return r.db.WithContext(ctx).Save(forecast).Error
}

// GetForecastBySKU 根据SKU ID获取销售预测记录。
// 如果记录未找到，则返回nil而非错误，由应用层进行判断。
func (r *inventoryForecastRepository) GetForecastBySKU(ctx context.Context, skuID uint64) (*entity.SalesForecast, error) {
	var forecast entity.SalesForecast
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).First(&forecast).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &forecast, nil
}

// --- 库存预警 (InventoryWarning methods) ---

// SaveWarning 将库存预警实体保存到数据库。
func (r *inventoryForecastRepository) SaveWarning(ctx context.Context, warning *entity.InventoryWarning) error {
	return r.db.WithContext(ctx).Save(warning).Error
}

// ListWarnings 从数据库列出所有库存预警记录，支持分页。
func (r *inventoryForecastRepository) ListWarnings(ctx context.Context, offset, limit int) ([]*entity.InventoryWarning, int64, error) {
	var list []*entity.InventoryWarning
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.InventoryWarning{})

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 滞销品 (SlowMovingItem methods) ---

// SaveSlowMovingItem 将滞销品实体保存到数据库。
func (r *inventoryForecastRepository) SaveSlowMovingItem(ctx context.Context, item *entity.SlowMovingItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// ListSlowMovingItems 从数据库列出所有滞销品记录，支持分页。
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

// --- 补货建议 (ReplenishmentSuggestion methods) ---

// SaveReplenishmentSuggestion 将补货建议实体保存到数据库。
func (r *inventoryForecastRepository) SaveReplenishmentSuggestion(ctx context.Context, suggestion *entity.ReplenishmentSuggestion) error {
	return r.db.WithContext(ctx).Save(suggestion).Error
}

// ListReplenishmentSuggestions 从数据库列出所有补货建议记录，支持通过优先级过滤和分页。
func (r *inventoryForecastRepository) ListReplenishmentSuggestions(ctx context.Context, priority string, offset, limit int) ([]*entity.ReplenishmentSuggestion, int64, error) {
	var list []*entity.ReplenishmentSuggestion
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.ReplenishmentSuggestion{})
	if priority != "" { // 如果提供了优先级，则按优先级过滤。
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

// --- 缺货风险 (StockoutRisk methods) ---

// SaveStockoutRisk 将缺货风险实体保存到数据库。
func (r *inventoryForecastRepository) SaveStockoutRisk(ctx context.Context, risk *entity.StockoutRisk) error {
	return r.db.WithContext(ctx).Save(risk).Error
}

// ListStockoutRisks 从数据库列出所有缺货风险记录，支持通过风险等级过滤和分页。
func (r *inventoryForecastRepository) ListStockoutRisks(ctx context.Context, level entity.StockoutRiskLevel, offset, limit int) ([]*entity.StockoutRisk, int64, error) {
	var list []*entity.StockoutRisk
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.StockoutRisk{})
	if level != "" { // 如果提供了风险等级，则按风险等级过滤。
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

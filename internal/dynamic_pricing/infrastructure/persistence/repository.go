package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/domain/entity"     // 导入动态定价模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/domain/repository" // 导入动态定价模块的领域仓储接口。
	"time"                                                                      // 导入时间包，用于查询条件。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// pricingRepository 是 PricingRepository 接口的GORM实现。
// 它负责将动态定价模块的领域实体映射到数据库，并执行持久化操作。
type pricingRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewPricingRepository 创建并返回一个新的 pricingRepository 实例。
// db: GORM数据库连接实例。
func NewPricingRepository(db *gorm.DB) repository.PricingRepository {
	return &pricingRepository{db: db}
}

// --- DynamicPrice methods ---

// SaveDynamicPrice 将动态价格实体保存到数据库。
// 如果动态价格已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *pricingRepository) SaveDynamicPrice(ctx context.Context, price *entity.DynamicPrice) error {
	return r.db.WithContext(ctx).Save(price).Error
}

// GetLatestDynamicPrice 获取指定SKU的最新动态价格记录。
func (r *pricingRepository) GetLatestDynamicPrice(ctx context.Context, skuID uint64) (*entity.DynamicPrice, error) {
	var price entity.DynamicPrice
	// 按创建时间降序排序，取第一条记录，即为最新价格。
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).Order("created_at desc").First(&price).Error; err != nil {
		return nil, err
	}
	return &price, nil
}

// ListDynamicPrices 从数据库列出指定SKU的所有历史动态价格记录，支持分页。
func (r *pricingRepository) ListDynamicPrices(ctx context.Context, skuID uint64, offset, limit int) ([]*entity.DynamicPrice, int64, error) {
	var list []*entity.DynamicPrice
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.DynamicPrice{})
	if skuID != 0 { // 如果提供了SKUID，则按SKUID过滤。
		db = db.Where("sku_id = ?", skuID)
	}

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

// --- CompetitorPrice methods ---

// SaveCompetitorPriceInfo 将竞品价格汇总信息实体保存到数据库。
// 该方法在一个事务中保存主实体及其关联的竞品价格列表。
func (r *pricingRepository) SaveCompetitorPriceInfo(ctx context.Context, info *entity.CompetitorPriceInfo) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 保存CompetitorPriceInfo主实体。
		if err := tx.Save(info).Error; err != nil {
			return err
		}
		// 遍历并保存所有关联的CompetitorPrice实体。
		for _, comp := range info.Competitors {
			comp.InfoID = uint64(info.ID) // 确保关联的外键已设置。
			if err := tx.Save(comp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetCompetitorPriceInfo 获取指定SKU的竞品价格汇总信息实体，并预加载其关联的竞品价格列表。
func (r *pricingRepository) GetCompetitorPriceInfo(ctx context.Context, skuID uint64) (*entity.CompetitorPriceInfo, error) {
	var info entity.CompetitorPriceInfo
	// Preload "Competitors" 确保在获取CompetitorPriceInfo时，同时加载所有关联的CompetitorPrice。
	if err := r.db.WithContext(ctx).Preload("Competitors").Where("sku_id = ?", skuID).Order("created_at desc").First(&info).Error; err != nil {
		return nil, err
	}
	return &info, nil
}

// --- PriceHistory methods ---

// SavePriceHistory 将价格历史数据实体保存到数据库。
func (r *pricingRepository) SavePriceHistory(ctx context.Context, history *entity.PriceHistoryData) error {
	return r.db.WithContext(ctx).Save(history).Error
}

// GetPriceHistory 获取指定SKU在过去指定天数内的价格历史数据。
func (r *pricingRepository) GetPriceHistory(ctx context.Context, skuID uint64, days int) ([]*entity.PriceHistoryData, error) {
	var list []*entity.PriceHistoryData
	startTime := time.Now().AddDate(0, 0, -days) // 计算起始时间。
	// 查询指定SKU且在时间范围内的价格历史数据，按日期升序排列。
	if err := r.db.WithContext(ctx).Where("sku_id = ? AND date >= ?", skuID, startTime).Order("date asc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- PricingStrategy methods ---

// SavePricingStrategy 将定价策略实体保存到数据库。
func (r *pricingRepository) SavePricingStrategy(ctx context.Context, strategy *entity.PricingStrategy) error {
	return r.db.WithContext(ctx).Save(strategy).Error
}

// GetPricingStrategy 获取指定SKU的定价策略实体。
func (r *pricingRepository) GetPricingStrategy(ctx context.Context, skuID uint64) (*entity.PricingStrategy, error) {
	var strategy entity.PricingStrategy
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).First(&strategy).Error; err != nil {
		return nil, err
	}
	return &strategy, nil
}

// ListPricingStrategies 从数据库列出所有定价策略记录，支持分页。
func (r *pricingRepository) ListPricingStrategies(ctx context.Context, offset, limit int) ([]*entity.PricingStrategy, int64, error) {
	var list []*entity.PricingStrategy
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.PricingStrategy{})

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

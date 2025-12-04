package persistence

import (
	"context"
	"errors"
	"github.com/wyfcoding/ecommerce/internal/pricing/domain/entity"     // 导入定价模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/pricing/domain/repository" // 导入定价模块的领域仓储接口。
	"time"                                                              // 导入时间包，用于查询条件。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type pricingRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewPricingRepository 创建并返回一个新的 pricingRepository 实例。
func NewPricingRepository(db *gorm.DB) repository.PricingRepository {
	return &pricingRepository{db: db}
}

// --- 规则 (PricingRule methods) ---

// SaveRule 将定价规则实体保存到数据库。
// 如果实体已存在，则更新；如果不存在，则创建。
func (r *pricingRepository) SaveRule(ctx context.Context, rule *entity.PricingRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

// GetRule 根据ID从数据库获取定价规则记录。
// 如果记录未找到，则返回nil。
func (r *pricingRepository) GetRule(ctx context.Context, id uint64) (*entity.PricingRule, error) {
	var rule entity.PricingRule
	if err := r.db.WithContext(ctx).First(&rule, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &rule, nil
}

// GetActiveRule 根据商品ID和SKUID获取当前活跃的定价规则记录。
// 活跃规则需要满足：已启用、当前时间在生效时间范围内，并按更新时间降序排列取第一条。
func (r *pricingRepository) GetActiveRule(ctx context.Context, productID, skuID uint64) (*entity.PricingRule, error) {
	var rule entity.PricingRule
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND sku_id = ? AND enabled = ? AND start_time <= ? AND end_time >= ?",
						productID, skuID, true, now, now). // 过滤条件：商品ID、SKUID、已启用、生效时间范围。
		Order("updated_at desc"). // 按更新时间倒序排列，取最新更新的活跃规则。
		First(&rule).Error        // 取第一条记录。
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &rule, nil
}

// ListRules 从数据库列出指定商品ID的所有定价规则记录，支持分页。
func (r *pricingRepository) ListRules(ctx context.Context, productID uint64, offset, limit int) ([]*entity.PricingRule, int64, error) {
	var list []*entity.PricingRule
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.PricingRule{})
	if productID > 0 { // 如果提供了商品ID，则按商品ID过滤。
		db = db.Where("product_id = ?", productID)
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

// --- 历史 (PriceHistory methods) ---

// SaveHistory 将价格历史实体保存到数据库。
func (r *pricingRepository) SaveHistory(ctx context.Context, history *entity.PriceHistory) error {
	return r.db.WithContext(ctx).Save(history).Error
}

// ListHistory 从数据库列出指定商品ID和SKUID的所有价格历史记录，支持分页。
func (r *pricingRepository) ListHistory(ctx context.Context, productID, skuID uint64, offset, limit int) ([]*entity.PriceHistory, int64, error) {
	var list []*entity.PriceHistory
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.PriceHistory{})
	if productID > 0 { // 如果提供了商品ID，则按商品ID过滤。
		db = db.Where("product_id = ?", productID)
	}
	if skuID > 0 { // 如果提供了SKUID，则按SKUID过滤。
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

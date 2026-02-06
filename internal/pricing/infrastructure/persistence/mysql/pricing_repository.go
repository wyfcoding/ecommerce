package mysql

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

// NewPricingRepository 创建并返回一个新的 PricingRepository 实例。
func NewPricingRepository(db *gorm.DB) domain.PricingRepository {
	return &pricingRepository{db: db}
}

func (r *pricingRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *pricingRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *pricingRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *pricingRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- PricingRule ---

func (r *pricingRepository) SaveRule(ctx context.Context, rule *domain.PricingRule) error {
	return r.saveRuleWithTx(ctx, r.db, rule)
}

func (r *pricingRepository) SaveRuleInTx(ctx context.Context, tx any, rule *domain.PricingRule) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveRuleWithTx(ctx, gormTx, rule)
}

func (r *pricingRepository) GetRule(ctx context.Context, id uint64) (*domain.PricingRule, error) {
	var model PricingRuleModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toPricingRule(&model), nil
}

func (r *pricingRepository) GetActiveRule(ctx context.Context, productID, skuID uint64) (*domain.PricingRule, error) {
	var model PricingRuleModel
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND sku_id = ? AND enabled = ? AND start_time <= ? AND end_time >= ?",
			productID, skuID, true, now, now).
		Order("updated_at desc").
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toPricingRule(&model), nil
}

func (r *pricingRepository) ListRules(ctx context.Context, productID uint64, offset, limit int) ([]*domain.PricingRule, int64, error) {
	var list []*PricingRuleModel
	var total int64

	db := r.db.WithContext(ctx).Model(&PricingRuleModel{})
	if productID > 0 {
		db = db.Where("product_id = ?", productID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.PricingRule, len(list))
	for i, model := range list {
		items[i] = toPricingRule(model)
	}

	return items, total, nil
}

// --- PriceHistory ---

func (r *pricingRepository) SaveHistory(ctx context.Context, history *domain.PriceHistory) error {
	return r.saveHistoryWithTx(ctx, r.db, history)
}

func (r *pricingRepository) SaveHistoryInTx(ctx context.Context, tx any, history *domain.PriceHistory) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveHistoryWithTx(ctx, gormTx, history)
}

func (r *pricingRepository) GetHistory(ctx context.Context, id uint64) (*domain.PriceHistory, error) {
	var model PriceHistoryModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toPriceHistory(&model), nil
}

func (r *pricingRepository) ListHistory(ctx context.Context, productID, skuID uint64, offset, limit int) ([]*domain.PriceHistory, int64, error) {
	var list []*PriceHistoryModel
	var total int64

	db := r.db.WithContext(ctx).Model(&PriceHistoryModel{})
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

	items := make([]*domain.PriceHistory, len(list))
	for i, model := range list {
		items[i] = toPriceHistory(model)
	}

	return items, total, nil
}

func (r *pricingRepository) saveRuleWithTx(ctx context.Context, tx *gorm.DB, rule *domain.PricingRule) error {
	if rule == nil {
		return nil
	}
	model := toPricingRuleModel(rule)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	rule.ID = uint64(model.ID)
	rule.CreatedAt = model.CreatedAt
	rule.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *pricingRepository) saveHistoryWithTx(ctx context.Context, tx *gorm.DB, history *domain.PriceHistory) error {
	if history == nil {
		return nil
	}
	model := toPriceHistoryModel(history)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	history.ID = uint64(model.ID)
	history.CreatedAt = model.CreatedAt
	history.UpdatedAt = model.UpdatedAt
	return nil
}

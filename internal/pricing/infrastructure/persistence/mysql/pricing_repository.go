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

// --- DynamicPricing ---

func (r *pricingRepository) SaveDynamicPrice(ctx context.Context, price *domain.DynamicPrice) error {
	return r.saveDynamicPriceWithTx(ctx, r.db, price)
}

func (r *pricingRepository) SaveDynamicPriceInTx(ctx context.Context, tx any, price *domain.DynamicPrice) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveDynamicPriceWithTx(ctx, gormTx, price)
}

func (r *pricingRepository) GetLatestDynamicPrice(ctx context.Context, skuID uint64) (*domain.DynamicPrice, error) {
	var model DynamicPriceModel
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).Order("created_at desc").First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDynamicPrice(&model), nil
}

// --- Competitor & Elasticity ---

func (r *pricingRepository) SaveCompetitorPrice(ctx context.Context, price *domain.CompetitorPrice) error {
	model := toCompetitorPriceModel(price)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	price.ID = model.ID
	return nil
}

func (r *pricingRepository) SaveCompetitorPriceInTx(ctx context.Context, tx any, price *domain.CompetitorPrice) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid tx")
	}
	model := toCompetitorPriceModel(price)
	if err := gormTx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	price.ID = model.ID
	return nil
}

func (r *pricingRepository) SaveCompetitorPriceInfo(ctx context.Context, info *domain.CompetitorPriceInfo) error {
	model := toCompetitorPriceInfoModel(info)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	info.ID = model.ID
	return nil
}

func (r *pricingRepository) SaveCompetitorPriceInfoInTx(ctx context.Context, tx any, info *domain.CompetitorPriceInfo) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid tx")
	}
	model := toCompetitorPriceInfoModel(info)
	if err := gormTx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	info.ID = model.ID
	return nil
}

func (r *pricingRepository) GetCompetitorPriceInfo(ctx context.Context, skuID uint64) (*domain.CompetitorPriceInfo, error) {
	var model CompetitorPriceInfoModel
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	info := toCompetitorPriceInfo(&model)
	var competitorModels []*CompetitorPriceModel
	if err := r.db.WithContext(ctx).Where("info_id = ?", info.ID).Find(&competitorModels).Error; err == nil {
		for _, cm := range competitorModels {
			info.Competitors = append(info.Competitors, toCompetitorPrice(cm))
		}
	}

	return info, nil
}

func (r *pricingRepository) GetPriceElasticity(ctx context.Context, skuID uint64) (*domain.PriceElasticity, error) {
	var model PriceElasticityModel
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).Order("analyzed_at desc").First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toPriceElasticity(&model), nil
}

func (r *pricingRepository) GetDynamicPriceHistory(ctx context.Context, skuID uint64, limit int) ([]domain.DynamicPrice, error) {
	var models []DynamicPriceModel
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).Order("created_at desc").Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}

	res := make([]domain.DynamicPrice, len(models))
	for i, m := range models {
		res[i] = *toDynamicPrice(&m)
	}
	return res, nil
}

// --- Helpers ---

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

func (r *pricingRepository) saveDynamicPriceWithTx(ctx context.Context, tx *gorm.DB, price *domain.DynamicPrice) error {
	if price == nil {
		return nil
	}
	model := toDynamicPriceModel(price)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	price.ID = model.ID
	price.CreatedAt = model.CreatedAt
	price.UpdatedAt = model.UpdatedAt
	return nil
}

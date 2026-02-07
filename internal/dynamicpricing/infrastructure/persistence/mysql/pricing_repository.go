package mysql

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

// --- tx helpers ---

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

// --- DynamicPrice methods ---

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
	var price DynamicPriceModel
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).Order("created_at desc").First(&price).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDynamicPrice(&price), nil
}

// --- CompetitorPrice methods ---

func (r *pricingRepository) SaveCompetitorPrice(ctx context.Context, price *domain.CompetitorPrice) error {
	return r.saveCompetitorPriceWithTx(ctx, r.db, price)
}

func (r *pricingRepository) SaveCompetitorPriceInTx(ctx context.Context, tx any, price *domain.CompetitorPrice) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveCompetitorPriceWithTx(ctx, gormTx, price)
}

func (r *pricingRepository) GetCompetitorPriceInfo(ctx context.Context, skuID uint64) (*domain.CompetitorPriceInfo, error) {
	var info CompetitorPriceInfoModel
	if err := r.db.WithContext(ctx).Preload("Competitors").Where("sku_id = ?", skuID).Order("created_at desc").First(&info).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toCompetitorPriceInfo(&info), nil
}

func (r *pricingRepository) SaveCompetitorPriceInfo(ctx context.Context, info *domain.CompetitorPriceInfo) error {
	return r.saveCompetitorPriceInfoWithTx(ctx, r.db, info)
}

func (r *pricingRepository) SaveCompetitorPriceInfoInTx(ctx context.Context, tx any, info *domain.CompetitorPriceInfo) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveCompetitorPriceInfoWithTx(ctx, gormTx, info)
}

// --- PriceHistory methods ---

func (r *pricingRepository) GetPriceHistory(ctx context.Context, skuID uint64, limit int) ([]domain.PriceHistoryData, error) {
	var list []PriceHistoryDataModel
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).Order("date desc").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return toPriceHistoryData(list), nil
}

// --- PriceElasticity methods ---

func (r *pricingRepository) GetPriceElasticity(ctx context.Context, skuID uint64) (*domain.PriceElasticity, error) {
	var elasticity PriceElasticityModel
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).Order("created_at desc").First(&elasticity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toPriceElasticity(&elasticity), nil
}

// --- PricingStrategy methods ---

func (r *pricingRepository) SavePricingStrategy(ctx context.Context, strategy *domain.PricingStrategy) error {
	return r.savePricingStrategyWithTx(ctx, r.db, strategy)
}

func (r *pricingRepository) SavePricingStrategyInTx(ctx context.Context, tx any, strategy *domain.PricingStrategy) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.savePricingStrategyWithTx(ctx, gormTx, strategy)
}

func (r *pricingRepository) GetPricingStrategy(ctx context.Context, skuID uint64) (*domain.PricingStrategy, error) {
	var strategy PricingStrategyModel
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).First(&strategy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toPricingStrategy(&strategy), nil
}

func (r *pricingRepository) ListPricingStrategies(ctx context.Context, query *domain.PricingStrategyQuery) ([]*domain.PricingStrategy, int64, error) {
	var list []*PricingStrategyModel
	var total int64

	db := r.db.WithContext(ctx).Model(&PricingStrategyModel{})
	if query != nil {
		if query.SKUID > 0 {
			db = db.Where("sku_id = ?", query.SKUID)
		}
		if query.Enabled != nil {
			db = db.Where("enabled = ?", *query.Enabled)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := 1
	pageSize := 20
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.PricingStrategy, len(list))
	for i, item := range list {
		items[i] = toPricingStrategy(item)
	}
	return items, total, nil
}

func (r *pricingRepository) saveDynamicPriceWithTx(ctx context.Context, tx *gorm.DB, price *domain.DynamicPrice) error {
	model := toDynamicPriceModel(price)
	if model == nil {
		return nil
	}
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	price.ID = model.ID
	price.CreatedAt = model.CreatedAt
	price.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *pricingRepository) savePricingStrategyWithTx(ctx context.Context, tx *gorm.DB, strategy *domain.PricingStrategy) error {
	model := toPricingStrategyModel(strategy)
	if model == nil {
		return nil
	}
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	strategy.ID = model.ID
	strategy.CreatedAt = model.CreatedAt
	strategy.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *pricingRepository) saveCompetitorPriceWithTx(ctx context.Context, tx *gorm.DB, price *domain.CompetitorPrice) error {
	model := toCompetitorPriceModel(price)
	if model == nil {
		return nil
	}
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	price.ID = model.ID
	price.CreatedAt = model.CreatedAt
	price.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *pricingRepository) saveCompetitorPriceInfoWithTx(ctx context.Context, tx *gorm.DB, info *domain.CompetitorPriceInfo) error {
	model := toCompetitorPriceInfoModel(info)
	if model == nil {
		return nil
	}
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	info.ID = model.ID
	info.CreatedAt = model.CreatedAt
	info.UpdatedAt = model.UpdatedAt
	return nil
}

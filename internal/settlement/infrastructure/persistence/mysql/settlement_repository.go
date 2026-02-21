package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
	"gorm.io/gorm"
)

type settlementRepository struct {
	db *gorm.DB
}

// NewSettlementRepository 创建并返回一个新的 SettlementRepository 实例。
func NewSettlementRepository(db *gorm.DB) domain.SettlementRepository {
	return &settlementRepository{db: db}
}

// BeginTx 开始事务
func (r *settlementRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

// CommitTx 提交事务
func (r *settlementRepository) CommitTx(tx any) error {
	return tx.(*gorm.DB).Commit().Error
}

// RollbackTx 回滚事务
func (r *settlementRepository) RollbackTx(tx any) error {
	return tx.(*gorm.DB).Rollback().Error
}

// WithTx 事务包装器
func (r *settlementRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- 结算单管理 (Settlement methods) ---

func (r *settlementRepository) SaveSettlement(ctx context.Context, settlement *domain.Settlement) error {
	model := toSettlementModel(settlement)
	if model == nil {
		return nil
	}
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	*settlement = *toSettlement(model)
	return nil
}

func (r *settlementRepository) SaveSettlementInTx(ctx context.Context, tx any, settlement *domain.Settlement) error {
	model := toSettlementModel(settlement)
	if model == nil {
		return nil
	}
	if err := tx.(*gorm.DB).WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	*settlement = *toSettlement(model)
	return nil
}

func (r *settlementRepository) GetSettlement(ctx context.Context, id uint64) (*domain.Settlement, error) {
	var model SettlementModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toSettlement(&model), nil
}

func (r *settlementRepository) GetSettlementByNo(ctx context.Context, no string) (*domain.Settlement, error) {
	var model SettlementModel
	if err := r.db.WithContext(ctx).Where("settlement_no = ?", no).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toSettlement(&model), nil
}

func (r *settlementRepository) Save(ctx context.Context, settlement *domain.Settlement) error {
	return r.SaveSettlement(ctx, settlement)
}

func (r *settlementRepository) Update(ctx context.Context, settlement *domain.Settlement) error {
	return r.SaveSettlement(ctx, settlement) // Save handles update in GORM
}

func (r *settlementRepository) GetByID(ctx context.Context, id string) (*domain.Settlement, error) {
	var model SettlementModel
	if err := r.db.WithContext(ctx).Where("settlement_no = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toSettlement(&model), nil
}

func (r *settlementRepository) GetByIDForUpdate(ctx context.Context, id string) (*domain.Settlement, error) {
	var model SettlementModel
	if err := r.db.WithContext(ctx).Clauses(nil).Where("settlement_no = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toSettlement(&model), nil
}

func (r *settlementRepository) GetByMerchantAndPeriod(ctx context.Context, merchantID uint64, start, end time.Time) (*domain.Settlement, error) {
	var model SettlementModel
	if err := r.db.WithContext(ctx).Where("merchant_id = ? AND start_date = ? AND end_date = ?", merchantID, start, end).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toSettlement(&model), nil
}

func (r *settlementRepository) ListByMerchant(ctx context.Context, merchantID uint64, status *domain.SettlementStatus, offset, limit int) ([]*domain.Settlement, int64, error) {
	return r.ListSettlements(ctx, merchantID, status, offset, limit)
}

func (r *settlementRepository) ListSettlements(ctx context.Context, merchantID uint64, status *domain.SettlementStatus, offset, limit int) ([]*domain.Settlement, int64, error) {
	var list []*SettlementModel
	var total int64

	db := r.db.WithContext(ctx).Model(&SettlementModel{})
	if merchantID > 0 {
		db = db.Where("merchant_id = ?", merchantID)
	}
	if status != nil && *status != "" {
		db = db.Where("status = ?", *status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	results := make([]*domain.Settlement, 0, len(list))
	for _, model := range list {
		results = append(results, toSettlement(model))
	}

	return results, total, nil
}

// --- 结算明细管理 (SettlementDetail methods) ---

func (r *settlementRepository) SaveSettlementDetail(ctx context.Context, detail *domain.SettlementDetail) error {
	model := toSettlementDetailModel(detail)
	if model == nil {
		return nil
	}
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	*detail = *toSettlementDetail(model)
	return nil
}

func (r *settlementRepository) SaveSettlementDetailInTx(ctx context.Context, tx any, detail *domain.SettlementDetail) error {
	model := toSettlementDetailModel(detail)
	if model == nil {
		return nil
	}
	if err := tx.(*gorm.DB).WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	*detail = *toSettlementDetail(model)
	return nil
}

func (r *settlementRepository) SaveDetails(ctx context.Context, details []*domain.SettlementDetail) error {
	models := make([]*SettlementDetailModel, 0, len(details))
	for _, d := range details {
		models = append(models, toSettlementDetailModel(d))
	}
	return r.db.WithContext(ctx).Save(&models).Error
}

func (r *settlementRepository) GetDetailsBySettlementID(ctx context.Context, settlementID string) ([]*domain.SettlementDetail, error) {
	var list []*SettlementDetailModel
	if err := r.db.WithContext(ctx).Where("settlement_id = ?", settlementID).Find(&list).Error; err != nil {
		return nil, err
	}
	results := make([]*domain.SettlementDetail, 0, len(list))
	for _, model := range list {
		results = append(results, toSettlementDetail(model))
	}
	return results, nil
}

func (r *settlementRepository) ListSettlementDetails(ctx context.Context, settlementID uint64) ([]*domain.SettlementDetail, error) {
	var list []*SettlementDetailModel
	if err := r.db.WithContext(ctx).Where("settlement_id = ?", settlementID).Find(&list).Error; err != nil {
		return nil, err
	}
	results := make([]*domain.SettlementDetail, 0, len(list))
	for _, model := range list {
		results = append(results, toSettlementDetail(model))
	}
	return results, nil
}

// --- 商户账户管理 (MerchantAccount methods) ---

func (r *settlementRepository) GetMerchantAccount(ctx context.Context, merchantID uint64) (*domain.MerchantAccount, error) {
	var model MerchantAccountModel
	if err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toMerchantAccount(&model), nil
}

func (r *settlementRepository) SaveMerchantAccount(ctx context.Context, account *domain.MerchantAccount) error {
	model := toMerchantAccountModel(account)
	if model == nil {
		return nil
	}
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	*account = *toMerchantAccount(model)
	return nil
}

func (r *settlementRepository) SaveMerchantAccountInTx(ctx context.Context, tx any, account *domain.MerchantAccount) error {
	model := toMerchantAccountModel(account)
	if model == nil {
		return nil
	}
	if err := tx.(*gorm.DB).WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	*account = *toMerchantAccount(model)
	return nil
}

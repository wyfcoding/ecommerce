package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/settlement/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/settlement/domain/repository"
	"errors"

	"gorm.io/gorm"
)

type settlementRepository struct {
	db *gorm.DB
}

func NewSettlementRepository(db *gorm.DB) repository.SettlementRepository {
	return &settlementRepository{db: db}
}

// 结算单
func (r *settlementRepository) SaveSettlement(ctx context.Context, settlement *entity.Settlement) error {
	return r.db.WithContext(ctx).Save(settlement).Error
}

func (r *settlementRepository) GetSettlement(ctx context.Context, id uint64) (*entity.Settlement, error) {
	var settlement entity.Settlement
	if err := r.db.WithContext(ctx).First(&settlement, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &settlement, nil
}

func (r *settlementRepository) GetSettlementByNo(ctx context.Context, no string) (*entity.Settlement, error) {
	var settlement entity.Settlement
	if err := r.db.WithContext(ctx).Where("settlement_no = ?", no).First(&settlement).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &settlement, nil
}

func (r *settlementRepository) ListSettlements(ctx context.Context, merchantID uint64, status *entity.SettlementStatus, offset, limit int) ([]*entity.Settlement, int64, error) {
	var list []*entity.Settlement
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Settlement{})
	if merchantID > 0 {
		db = db.Where("merchant_id = ?", merchantID)
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// 结算明细
func (r *settlementRepository) SaveSettlementDetail(ctx context.Context, detail *entity.SettlementDetail) error {
	return r.db.WithContext(ctx).Save(detail).Error
}

func (r *settlementRepository) ListSettlementDetails(ctx context.Context, settlementID uint64) ([]*entity.SettlementDetail, error) {
	var list []*entity.SettlementDetail
	if err := r.db.WithContext(ctx).Where("settlement_id = ?", settlementID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// 商户账户
func (r *settlementRepository) GetMerchantAccount(ctx context.Context, merchantID uint64) (*entity.MerchantAccount, error) {
	var account entity.MerchantAccount
	if err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *settlementRepository) SaveMerchantAccount(ctx context.Context, account *entity.MerchantAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

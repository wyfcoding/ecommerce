package infrastructure

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/merchantsettlement/domain"
	"gorm.io/gorm"
)

type GormSettlementRepository struct {
	db *gorm.DB
}

func NewGormSettlementRepository(db *gorm.DB) *GormSettlementRepository {
	return &GormSettlementRepository{db: db}
}

func (r *GormSettlementRepository) Save(ctx context.Context, s *domain.Settlement) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *GormSettlementRepository) GetByID(ctx context.Context, id string) (*domain.Settlement, error) {
	var s domain.Settlement
	err := r.db.WithContext(ctx).Where("settlement_id = ?", id).First(&s).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &s, err
}

func (r *GormSettlementRepository) ListByMerchant(ctx context.Context, merchantID string, status string) ([]*domain.Settlement, error) {
	var ss []*domain.Settlement
	query := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&ss).Error
	return ss, err
}

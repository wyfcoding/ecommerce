package infrastructure

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/procurement/domain"
	"gorm.io/gorm"
)

type ProcurementRepository struct {
	db *gorm.DB
}

func NewProcurementRepository(db *gorm.DB) *ProcurementRepository {
	return &ProcurementRepository{db: db}
}

func (r *ProcurementRepository) SavePurchaseRequest(ctx context.Context, pr *domain.PurchaseRequest) error {
	return r.db.WithContext(ctx).Save(pr).Error
}

func (r *ProcurementRepository) GetPurchaseRequest(ctx context.Context, id string) (*domain.PurchaseRequest, error) {
	var pr domain.PurchaseRequest
	if err := r.db.WithContext(ctx).Preload("Items").Where("request_id = ?", id).First(&pr).Error; err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *ProcurementRepository) SavePurchaseOrder(ctx context.Context, po *domain.PurchaseOrder) error {
	return r.db.WithContext(ctx).Save(po).Error
}

func (r *ProcurementRepository) GetPurchaseOrder(ctx context.Context, id string) (*domain.PurchaseOrder, error) {
	var po domain.PurchaseOrder
	if err := r.db.WithContext(ctx).Preload("Items").Where("order_id = ?", id).First(&po).Error; err != nil {
		return nil, err
	}
	return &po, nil
}

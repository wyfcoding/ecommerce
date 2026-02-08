package infrastructure

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/supplier/domain"
	"gorm.io/gorm"
)

type SupplierRepository struct {
	db *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

func (r *SupplierRepository) Save(ctx context.Context, s *domain.Supplier) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *SupplierRepository) Get(ctx context.Context, id string) (*domain.Supplier, error) {
	var s domain.Supplier
	if err := r.db.WithContext(ctx).Preload("Supplies").Where("supplier_id = ?", id).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

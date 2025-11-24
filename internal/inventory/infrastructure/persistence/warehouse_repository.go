package persistence

import (
	"context"
	"ecommerce/internal/inventory/domain/entity"
	"ecommerce/internal/inventory/domain/repository"

	"gorm.io/gorm"
)

type warehouseRepository struct {
	db *gorm.DB
}

func NewWarehouseRepository(db *gorm.DB) repository.WarehouseRepository {
	return &warehouseRepository{db: db}
}

func (r *warehouseRepository) Save(ctx context.Context, warehouse *entity.Warehouse) error {
	return r.db.WithContext(ctx).Save(warehouse).Error
}

func (r *warehouseRepository) GetByID(ctx context.Context, id uint64) (*entity.Warehouse, error) {
	var warehouse entity.Warehouse
	if err := r.db.WithContext(ctx).First(&warehouse, id).Error; err != nil {
		return nil, err
	}
	return &warehouse, nil
}

func (r *warehouseRepository) ListAll(ctx context.Context) ([]*entity.Warehouse, error) {
	var list []*entity.Warehouse
	if err := r.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

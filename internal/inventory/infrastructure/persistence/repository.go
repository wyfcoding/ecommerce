package persistence

import (
	"context"
	"ecommerce/internal/inventory/domain/entity"
	"ecommerce/internal/inventory/domain/repository"
	"errors"

	"gorm.io/gorm"
)

type inventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) repository.InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) Save(ctx context.Context, inventory *entity.Inventory) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(inventory).Error; err != nil {
			return err
		}
		for _, log := range inventory.Logs {
			if log.ID == 0 { // Only save new logs
				log.InventoryID = uint64(inventory.ID)
				if err := tx.Save(log).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *inventoryRepository) GetBySkuID(ctx context.Context, skuID uint64) (*entity.Inventory, error) {
	var inventory entity.Inventory
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).First(&inventory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if not found, let service handle
		}
		return nil, err
	}
	return &inventory, nil
}

func (r *inventoryRepository) List(ctx context.Context, offset, limit int) ([]*entity.Inventory, int64, error) {
	var list []*entity.Inventory
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Inventory{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *inventoryRepository) GetLogs(ctx context.Context, inventoryID uint64, offset, limit int) ([]*entity.InventoryLog, int64, error) {
	var list []*entity.InventoryLog
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.InventoryLog{}).Where("inventory_id = ?", inventoryID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
	"github.com/wyfcoding/pkg/databases/sharding"

	"gorm.io/gorm"
)

type inventoryRepository struct {
	sharding *sharding.Manager
}

// NewInventoryRepository 创建分片库存仓储。
func NewInventoryRepository(sharding *sharding.Manager) domain.InventoryRepository {
	return &inventoryRepository{sharding: sharding}
}

// Save 将库存实体保存到对应分片。
func (r *inventoryRepository) Save(ctx context.Context, inventory *domain.Inventory) error {
	db := r.sharding.GetDB(inventory.SkuID)
	return db.WithContext(ctx).Save(inventory).Error
}

// SaveWithOptimisticLock 使用乐观锁保存。
func (r *inventoryRepository) SaveWithOptimisticLock(ctx context.Context, inventory *domain.Inventory) error {
	db := r.sharding.GetDB(inventory.SkuID)
	if inventory.ID == 0 {
		return db.WithContext(ctx).Create(inventory).Error
	}

	currentVersion := inventory.Version
	inventory.Version++

	res := db.WithContext(ctx).Model(inventory).
		Where("id = ? AND version = ?", inventory.ID, currentVersion).
		Updates(inventory)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("optimistic lock failed")
	}
	return nil
}

// SaveLog 保存库存日志到对应分片。
func (r *inventoryRepository) SaveLog(ctx context.Context, log *domain.InventoryLog) error {
	db := r.sharding.GetDB(log.SkuID)
	return db.WithContext(ctx).Create(log).Error
}

// GetBySkuID 定向查询分片。
func (r *inventoryRepository) GetBySkuID(ctx context.Context, skuID uint64) (*domain.Inventory, error) {
	db := r.sharding.GetDB(skuID)
	var inventory domain.Inventory
	if err := db.WithContext(ctx).Where("sku_id = ?", skuID).First(&inventory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &inventory, nil
}

// GetBySkuIDs 跨分片查询。
func (r *inventoryRepository) GetBySkuIDs(ctx context.Context, skuIDs []uint64) ([]*domain.Inventory, error) {
	var allList []*domain.Inventory
	for _, id := range skuIDs {
		inv, err := r.GetBySkuID(ctx, id)
		if err == nil && inv != nil {
			allList = append(allList, inv)
		}
	}
	return allList, nil
}

// List 扫描所有分片。
func (r *inventoryRepository) List(ctx context.Context, offset, limit int) ([]*domain.Inventory, int64, error) {
	dbs := r.sharding.GetAllDBs()
	var allList []*domain.Inventory
	var totalCount int64

	for _, db := range dbs {
		var list []*domain.Inventory
		var count int64
		query := db.WithContext(ctx).Model(&domain.Inventory{})
		query.Count(&count)
		totalCount += count

		query.Offset(offset).Limit(limit).Order("created_at desc").Find(&list)
		allList = append(allList, list...)
	}

	if len(allList) > limit {
		allList = allList[:limit]
	}

	return allList, totalCount, nil
}

// Delete 从对应分片删除。
func (r *inventoryRepository) Delete(ctx context.Context, skuID uint64) error {
	db := r.sharding.GetDB(skuID)
	return db.WithContext(ctx).Where("sku_id = ?", skuID).Delete(&domain.Inventory{}).Error
}

// GetLogs 获取指定分片下的日志。
func (r *inventoryRepository) GetLogs(ctx context.Context, skuID uint64, inventoryID uint64, offset, limit int) ([]*domain.InventoryLog, int64, error) {
	db := r.sharding.GetDB(skuID)
	var list []*domain.InventoryLog
	var total int64

	query := db.WithContext(ctx).Model(&domain.InventoryLog{}).Where("inventory_id = ?", inventoryID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

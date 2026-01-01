package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/inventory/domain" // 导入库存模块的领域层。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type inventoryRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewInventoryRepository 创建并返回一个新的 inventoryRepository 实例。
// db: GORM数据库连接实例。
func NewInventoryRepository(db *gorm.DB) domain.InventoryRepository {
	return &inventoryRepository{db: db}
}

// Save 将库存实体保存到数据库。
func (r *inventoryRepository) Save(ctx context.Context, inventory *domain.Inventory) error {
	return r.db.WithContext(ctx).Save(inventory).Error
}

// SaveWithOptimisticLock 使用乐观锁保存库存实体。
func (r *inventoryRepository) SaveWithOptimisticLock(ctx context.Context, inventory *domain.Inventory) error {
	if inventory.ID == 0 {
		return r.Save(ctx, inventory)
	}

	currentVersion := inventory.Version
	inventory.Version++

	// 使用 Updates 更新所有字段，包括零值（如果需要，应使用 Select("*") 或指定字段）
	// 这里假设 inventory 包含了所有最新状态
	res := r.db.WithContext(ctx).Model(inventory).
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

// SaveLog 保存库存日志。
func (r *inventoryRepository) SaveLog(ctx context.Context, log *domain.InventoryLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetBySkuID 根据SKU ID从数据库获取库存记录。
// 如果记录未找到，则返回nil而非错误，由应用层进行判断。
func (r *inventoryRepository) GetBySkuID(ctx context.Context, skuID uint64) (*domain.Inventory, error) {
	var inventory domain.Inventory
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).First(&inventory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &inventory, nil
}

// GetBySkuIDs 根据SKU ID列表获取多个库存记录。
func (r *inventoryRepository) GetBySkuIDs(ctx context.Context, skuIDs []uint64) ([]*domain.Inventory, error) {
	var list []*domain.Inventory
	// 使用IN查询获取多个SKU的库存。
	if err := r.db.WithContext(ctx).Where("sku_id IN ?", skuIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// List 从数据库列出所有库存记录，支持分页。
func (r *inventoryRepository) List(ctx context.Context, offset, limit int) ([]*domain.Inventory, int64, error) {
	var list []*domain.Inventory
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Inventory{})

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// GetLogs 获取指定库存ID的所有库存日志，支持分页。
func (r *inventoryRepository) GetLogs(ctx context.Context, inventoryID uint64, offset, limit int) ([]*domain.InventoryLog, int64, error) {
	var list []*domain.InventoryLog
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.InventoryLog{}).Where("inventory_id = ?", inventoryID)

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

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
// 如果库存已存在（通过ID），则更新其信息；如果不存在，则创建。
// 此方法在一个事务中保存库存主实体及其关联的新增日志。
func (r *inventoryRepository) Save(ctx context.Context, inventory *domain.Inventory) error {
	// 使用事务确保库存主表和日志表的更新操作的原子性。
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 保存或更新库存主实体。
		if err := tx.Save(inventory).Error; err != nil {
			return err
		}
		// 遍历所有日志，只保存新增的日志（ID为0的日志）。
		for _, log := range inventory.Logs {
			if log.ID == 0 { // 检查是否是新日志。
				log.InventoryID = uint64(inventory.ID) // 关联库存ID。
				if err := tx.Save(log).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
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

package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"ecommerce/internal/inventory/model"
)

// InventoryRepository 定义了库存数据仓库的接口
type InventoryRepository interface {
	GetInventory(ctx context.Context, sku string, warehouseID uint) (*model.Inventory, error)
	// AdjustStock 是核心方法，处理所有库存变动
	AdjustStock(ctx context.Context, sku string, warehouseID uint, quantityChange int, movementType model.MovementType, reference, reason string) error
	// ReserveStock 预留库存
	ReserveStock(ctx context.Context, sku string, warehouseID uint, quantityToReserve int) error
	// ReleaseStock 释放预留库存
	ReleaseStock(ctx context.Context, sku string, warehouseID uint, quantityToRelease int) error
}

// inventoryRepository 是接口的具体实现
type inventoryRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

// NewInventoryRepository 创建一个新的 inventoryRepository 实例
func NewInventoryRepository(db *gorm.DB, cache *redis.Client) InventoryRepository {
	return &inventoryRepository{db: db, cache: cache}
}

// getLockKey 生成用于分布式锁的 Redis key
func (r *inventoryRepository) getLockKey(sku string, warehouseID uint) string {
	return fmt.Sprintf("lock:inventory:%s:%d", sku, warehouseID)
}

// GetInventory 获取指定 SKU 在特定仓库的库存信息
func (r *inventoryRepository) GetInventory(ctx context.Context, sku string, warehouseID uint) (*model.Inventory, error) {
	var inventory model.Inventory
	if err := r.db.WithContext(ctx).Where("product_sku = ? AND warehouse_id = ?", sku, warehouseID).First(&inventory).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 库存记录不存在是正常情况
		}
		return nil, fmt.Errorf("数据库查询库存失败: %w", err)
	}
	return &inventory, nil
}

// AdjustStock 原子地调整库存并记录流水
// 这是最通用的库存变更方法
func (r *inventoryRepository) AdjustStock(ctx context.Context, sku string, warehouseID uint, quantityChange int, movementType model.MovementType, reference, reason string) error {
	// 使用分布式锁确保对同一 SKU 的操作是串行的
	lockKey := r.getLockKey(sku, warehouseID)
	lock := r.cache.SetNX(ctx, lockKey, "locked", 10*time.Second) // 10秒超时防止死锁
	if !lock.Val() {
		return fmt.Errorf("获取库存锁失败，请重试")
	}
	defer r.cache.Del(ctx, lockKey) // 确保锁被释放

	// 在数据库事务中执行操作
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inventory model.Inventory
		// 1. 获取当前库存记录 (使用 FOR UPDATE 行锁)
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("product_sku = ? AND warehouse_id = ?", sku, warehouseID).First(&inventory).Error; err != nil {
			return fmt.Errorf("锁定并获取库存记录失败: %w", err)
		}

		// 2. 检查库存是否足够 (仅对出库和调整减少时)
		if quantityChange < 0 && inventory.QuantityOnHand+quantityChange < 0 {
			return fmt.Errorf("物理库存不足")
		}

		// 3. 更新库存数量
		inventory.QuantityOnHand += quantityChange
		inventory.QuantityAvailable = inventory.QuantityOnHand - inventory.QuantityReserved
		if err := tx.Save(&inventory).Error; err != nil {
			return fmt.Errorf("更新库存数量失败: %w", err)
		}

		// 4. 创建库存流水记录
		movement := model.StockMovement{
			InventoryID: inventory.ID,
			Type:        movementType,
			Quantity:    quantityChange,
			Reference:   reference,
			Reason:      reason,
		}
		if err := tx.Create(&movement).Error; err != nil {
			return fmt.Errorf("创建库存流水失败: %w", err)
		}

		return nil // 事务成功
	})
}

// ReserveStock 预留库存 (例如，用户下单后)
func (r *inventoryRepository) ReserveStock(ctx context.Context, sku string, warehouseID uint, quantityToReserve int) error {
	// ... 类似 AdjustStock 的分布式锁和事务逻辑 ...
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 获取库存并锁定
		var inventory model.Inventory
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("product_sku = ? AND warehouse_id = ?", sku, warehouseID).First(&inventory).Error; err != nil {
			return err
		}

		// 2. 检查可用库存是否足够
		if inventory.QuantityAvailable < quantityToReserve {
			return fmt.Errorf("可用库存不足")
		}

		// 3. 更新预留库存和可用库存
		inventory.QuantityReserved += quantityToReserve
		inventory.QuantityAvailable -= quantityToReserve
		return tx.Save(&inventory).Error
	})
}

// ReleaseStock 释放预留库存 (例如，订单取消或发货完成)
func (r *inventoryRepository) ReleaseStock(ctx context.Context, sku string, warehouseID uint, quantityToRelease int) error {
	// ... 类似逻辑 ...
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inventory model.Inventory
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("product_sku = ? AND warehouse_id = ?", sku, warehouseID).First(&inventory).Error; err != nil {
			return err
		}

		// 更新预留库存和可用库存
		inventory.QuantityReserved -= quantityToRelease
		inventory.QuantityAvailable += quantityToRelease
		return tx.Save(&inventory).Error
	})
}

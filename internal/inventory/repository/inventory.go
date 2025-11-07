package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ecommerce/internal/inventory/model"
	"ecommerce/pkg/lock"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type InventoryRepository interface {
	GetInventoryBySkuID(ctx context.Context, skuID uint64) (*model.Inventory, error)
	CreateInventory(ctx context.Context, inventory *model.Inventory) error
	UpdateInventory(ctx context.Context, inventory *model.Inventory) error
	DeductStock(ctx context.Context, inventory *model.Inventory, log *model.InventoryLog) error
	AddStock(ctx context.Context, inventory *model.Inventory, log *model.InventoryLog) error
	LockStock(ctx context.Context, inventory *model.Inventory, log *model.InventoryLog) error
	UnlockStock(ctx context.Context, inventory *model.Inventory, log *model.InventoryLog) error
	GetInventoryLogs(ctx context.Context, skuID uint64, limit int) ([]*model.InventoryLog, error)
}

type inventoryRepository struct {
	db    *gorm.DB
	redis *redis.Client
	lock  lock.DistributedLock
}

func NewInventoryRepository(db *gorm.DB, redis *redis.Client, lock lock.DistributedLock) InventoryRepository {
	return &inventoryRepository{
		db:    db,
		redis: redis,
		lock:  lock,
	}
}

func (r *inventoryRepository) GetInventoryBySkuID(ctx context.Context, skuID uint64) (*model.Inventory, error) {
	// 先从缓存获取
	cacheKey := fmt.Sprintf("inventory:sku:%d", skuID)
	if inv := r.getInventoryFromCache(ctx, cacheKey); inv != nil {
		return inv, nil
	}

	var inventory model.Inventory
	if err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).First(&inventory).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	r.cacheInventory(ctx, &inventory)
	return &inventory, nil
}

func (r *inventoryRepository) CreateInventory(ctx context.Context, inventory *model.Inventory) error {
	if err := r.db.WithContext(ctx).Create(inventory).Error; err != nil {
		return err
	}
	r.cacheInventory(ctx, inventory)
	return nil
}

func (r *inventoryRepository) UpdateInventory(ctx context.Context, inventory *model.Inventory) error {
	if err := r.db.WithContext(ctx).Save(inventory).Error; err != nil {
		return err
	}
	r.cacheInventory(ctx, inventory)
	return nil
}

// DeductStock 扣减库存（带分布式锁和事务）
func (r *inventoryRepository) DeductStock(ctx context.Context, inventory *model.Inventory, log *model.InventoryLog) error {
	lockKey := fmt.Sprintf("lock:inventory:%d", inventory.SkuID)
	
	// 获取分布式锁
	token, err := r.lock.Lock(ctx, lockKey, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer r.lock.Unlock(ctx, lockKey, token)

	// 开启事务
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新库存
		if err := tx.Save(inventory).Error; err != nil {
			return err
		}

		// 记录日志
		if err := tx.Create(log).Error; err != nil {
			return err
		}

		// 更新缓存
		r.cacheInventory(ctx, inventory)
		return nil
	})
}

// AddStock 增加库存（带分布式锁和事务）
func (r *inventoryRepository) AddStock(ctx context.Context, inventory *model.Inventory, log *model.InventoryLog) error {
	lockKey := fmt.Sprintf("lock:inventory:%d", inventory.SkuID)
	
	token, err := r.lock.Lock(ctx, lockKey, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer r.lock.Unlock(ctx, lockKey, token)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 如果是新记录，使用Create
		if inventory.ID == 0 {
			if err := tx.Create(inventory).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Save(inventory).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(log).Error; err != nil {
			return err
		}

		r.cacheInventory(ctx, inventory)
		return nil
	})
}

// LockStock 锁定库存（带分布式锁和事务）
func (r *inventoryRepository) LockStock(ctx context.Context, inventory *model.Inventory, log *model.InventoryLog) error {
	lockKey := fmt.Sprintf("lock:inventory:%d", inventory.SkuID)
	
	token, err := r.lock.Lock(ctx, lockKey, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer r.lock.Unlock(ctx, lockKey, token)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(inventory).Error; err != nil {
			return err
		}

		if err := tx.Create(log).Error; err != nil {
			return err
		}

		r.cacheInventory(ctx, inventory)
		return nil
	})
}

// UnlockStock 解锁库存（带分布式锁和事务）
func (r *inventoryRepository) UnlockStock(ctx context.Context, inventory *model.Inventory, log *model.InventoryLog) error {
	lockKey := fmt.Sprintf("lock:inventory:%d", inventory.SkuID)
	
	token, err := r.lock.Lock(ctx, lockKey, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer r.lock.Unlock(ctx, lockKey, token)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(inventory).Error; err != nil {
			return err
		}

		if err := tx.Create(log).Error; err != nil {
			return err
		}

		r.cacheInventory(ctx, inventory)
		return nil
	})
}

func (r *inventoryRepository) GetInventoryLogs(ctx context.Context, skuID uint64, limit int) ([]*model.InventoryLog, error) {
	var logs []*model.InventoryLog
	query := r.db.WithContext(ctx).Where("sku_id = ?", skuID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *inventoryRepository) cacheInventory(ctx context.Context, inventory *model.Inventory) {
	if r.redis == nil {
		return
	}
	data, _ := json.Marshal(inventory)
	cacheKey := fmt.Sprintf("inventory:sku:%d", inventory.SkuID)
	r.redis.Set(ctx, cacheKey, data, 10*time.Minute)
}

func (r *inventoryRepository) getInventoryFromCache(ctx context.Context, key string) *model.Inventory {
	if r.redis == nil {
		return nil
	}
	data, err := r.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil
	}
	var inventory model.Inventory
	if err := json.Unmarshal(data, &inventory); err != nil {
		return nil
	}
	return &inventory
}

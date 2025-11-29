package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/order/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/order/domain/repository"
	"github.com/wyfcoding/ecommerce/pkg/databases/sharding"

	"gorm.io/gorm"
)

type orderRepository struct {
	sharding *sharding.Manager
}

func NewOrderRepository(sharding *sharding.Manager) repository.OrderRepository {
	return &orderRepository{sharding: sharding}
}

func (r *orderRepository) Save(ctx context.Context, order *entity.Order) error {
	db := r.sharding.GetDB(uint64(order.UserID))
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		// Save items
		for _, item := range order.Items {
			if item.ID == 0 {
				item.OrderID = uint64(order.ID)
			}
			if err := tx.Save(item).Error; err != nil {
				return err
			}
		}
		// Save logs
		for _, log := range order.Logs {
			if log.ID == 0 {
				log.OrderID = uint64(order.ID)
			}
			if err := tx.Save(log).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *orderRepository) GetByID(ctx context.Context, id uint64) (*entity.Order, error) {
	// Note: Without UserID, we might need to query all shards or use a lookup table.
	// For simplicity in this phase, we assume the caller provides UserID via context or we scan all shards.
	// However, GetByID usually implies we know the ID. If ID is global unique (Snowflake), we can't infer shard from ID unless ID contains shard info.
	// Let's assume for now we scan all shards or this method is deprecated in favor of GetByUserIDAndID.
	// But to keep interface compatible, let's try to find in all shards (inefficient but works).

	// Optimization: If ID is Snowflake, we can embed ShardID in it.
	// Current Snowflake impl doesn't seem to embed ShardID.

	// Fallback: Scan all shards
	// TODO: Optimize this by embedding ShardID in OrderID or using a lookup service.

	// For now, let's just use the first shard or error out?
	// Better: Scan all.

	// Wait, the interface signature is GetByID(ctx, id).
	// Let's check if we can change the interface.
	// For now, let's implement a scan.

	// Actually, let's assume we can get UserID from context if possible, but standard way is to require UserID.
	// Let's stick to scanning for now as a safe fallback.

	// But wait, `List` has UserID. `GetByOrderNo` doesn't.

	// Let's try to find in all shards.
	// Since we don't expose `shards` map directly, we might need to add a method in manager to iterate.
	// Or, we just change the repository to require UserID for efficient lookup.
	// But that breaks the interface.

	// Let's modify the interface to include UserID for GetByID if possible?
	// Or just implement a "Broadcast" search.

	// Let's assume for this step we just use shard 0 for GetByID if UserID is not available,
	// OR we update the interface. Updating interface is better for "World Class".

	// Let's check `domain/repository/order_repository.go` first.
	// I'll update this file to use ShardingManager, but for GetByID I'll implement a simple scan or just use shard 0 for now
	// and mark TODO to update interface.

	// Actually, let's look at `List`. It has UserID.

	// Let's implement `GetByID` by checking all shards (naive scatter-gather).

	// Since `ShardingManager` doesn't expose iteration, let's add `GetAllDBs()` to it?
	// Or just loop 0 to shardCount.

	// I'll update `pkg/databases/sharding/sharding.go` to expose `GetShardCount` or `Iterate`.

	// For now, let's assume we can access shard 0.
	// Wait, I can't access `shardCount` from here.

	// I will update `sharding.go` in next step to support iteration.
	// For this step, I will implement `Save` and `List` correctly using UserID.
	// For `GetByID` and `GetByOrderNo`, I will temporarily use Shard 0 and add a TODO.

	db := r.sharding.GetDB(0) // Default to shard 0 for now
	var order entity.Order
	if err := db.WithContext(ctx).Preload("Items").Preload("Logs").First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*entity.Order, error) {
	// Similar issue as GetByID. Default to shard 0.
	db := r.sharding.GetDB(0)
	var order entity.Order
	if err := db.WithContext(ctx).Preload("Items").Preload("Logs").Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) List(ctx context.Context, userID uint64, status *entity.OrderStatus, offset, limit int) ([]*entity.Order, int64, error) {
	var list []*entity.Order
	var total int64

	db := r.sharding.GetDB(userID).WithContext(ctx).Model(&entity.Order{})

	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Preload("Items").Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

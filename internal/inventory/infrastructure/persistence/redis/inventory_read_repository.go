package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
)

const (
	inventoryDetailPrefix = "inventory:detail:"
)

// inventoryReadRepository 基于 Redis 的库存读模型仓储。
type inventoryReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewInventoryReadRepository 创建库存读模型仓储。
func NewInventoryReadRepository(client redis.UniversalClient, ttl time.Duration) domain.InventoryReadRepository {
	return &inventoryReadRepository{client: client, ttl: ttl}
}

// Save 保存或更新库存读模型。
func (r *inventoryReadRepository) Save(ctx context.Context, inventory *domain.Inventory) error {
	if inventory == nil {
		return nil
	}
	data, err := json.Marshal(inventory)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.skuKey(inventory.SkuID), data, r.ttl).Err()
}

// GetBySkuID 根据 SKU ID 获取读模型。
func (r *inventoryReadRepository) GetBySkuID(ctx context.Context, skuID uint64) (*domain.Inventory, error) {
	data, err := r.client.Get(ctx, r.skuKey(skuID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var inventory domain.Inventory
	if err := json.Unmarshal(data, &inventory); err != nil {
		return nil, err
	}
	return &inventory, nil
}

// Delete 删除读模型数据。
func (r *inventoryReadRepository) Delete(ctx context.Context, skuID uint64) error {
	return r.client.Del(ctx, r.skuKey(skuID)).Err()
}

func (r *inventoryReadRepository) skuKey(skuID uint64) string {
	return fmt.Sprintf("%s%d", inventoryDetailPrefix, skuID)
}

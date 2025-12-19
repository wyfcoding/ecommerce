package cache

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
)

// deductStockScript 嵌入了用于原子性扣减库存的Lua脚本。
//
//go:embed lua/deduct_stock.lua
var deductStockScript string

// revertStockScript 嵌入了用于原子性回滚库存的Lua脚本。
//
//go:embed lua/revert_stock.lua
var revertStockScript string

// RedisFlashSaleCache 结构体实现了 FlashSaleCache 接口，使用 Redis 作为底层缓存存储。
type RedisFlashSaleCache struct {
	client *redis.Client
}

// NewRedisFlashSaleCache 创建并返回一个新的 RedisFlashSaleCache 实例。
func NewRedisFlashSaleCache(client *redis.Client) *RedisFlashSaleCache {
	return &RedisFlashSaleCache{
		client: client,
	}
}

// SetStock 在Redis中设置指定秒杀活动的库存。
func (c *RedisFlashSaleCache) SetStock(ctx context.Context, flashsaleID uint64, stock int32) error {
	key := fmt.Sprintf("{flashsale:%d}:stock", flashsaleID)
	return c.client.Set(ctx, key, stock, 0).Err()
}

// DeductStock 在Redis中原子性地扣减指定秒杀活动的库存，并检查用户限购。
func (c *RedisFlashSaleCache) DeductStock(ctx context.Context, flashsaleID, userID uint64, quantity, limitPerUser int32) (bool, error) {
	keyTag := fmt.Sprintf("{flashsale:%d}", flashsaleID)
	res, err := c.client.Eval(ctx, deductStockScript, []string{keyTag}, userID, quantity, limitPerUser).Result()
	if err != nil {
		return false, err
	}

	code, ok := res.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected return type from lua script")
	}

	if code == 1 {
		return true, nil
	}
	return false, nil
}

// RevertStock 在Redis中原子性地回滚已扣减的库存。
func (c *RedisFlashSaleCache) RevertStock(ctx context.Context, flashsaleID, userID uint64, quantity int32) error {
	keyTag := fmt.Sprintf("{flashsale:%d}", flashsaleID)
	_, err := c.client.Eval(ctx, revertStockScript, []string{keyTag}, userID, quantity).Result()
	return err
}

// Ensure interface implementation 确保 RedisFlashSaleCache 实现了 domain.FlashSaleCache 接口。
var _ domain.FlashSaleCache = (*RedisFlashSaleCache)(nil)

package cache

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain/repository"
)

//go:embed lua/deduct_stock.lua
var deductStockScript string

//go:embed lua/revert_stock.lua
var revertStockScript string

type RedisFlashSaleCache struct {
	client *redis.Client
}

func NewRedisFlashSaleCache(client *redis.Client) *RedisFlashSaleCache {
	return &RedisFlashSaleCache{
		client: client,
	}
}

func (c *RedisFlashSaleCache) SetStock(ctx context.Context, flashsaleID uint64, stock int32) error {
	key := fmt.Sprintf("{flashsale:%d}:stock", flashsaleID)
	return c.client.Set(ctx, key, stock, 0).Err()
}

func (c *RedisFlashSaleCache) DeductStock(ctx context.Context, flashsaleID, userID uint64, quantity, limitPerUser int32) (bool, error) {
	keyTag := fmt.Sprintf("{flashsale:%d}", flashsaleID)
	// Keys: {flashsale:id}
	// Args: user_id, quantity, limit_per_user
	res, err := c.client.Eval(ctx, deductStockScript, []string{keyTag}, userID, quantity, limitPerUser).Result()
	if err != nil {
		return false, err
	}

	// Lua returns 1 for success, -1 for no stock, -2 for limit exceeded
	code, ok := res.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected return type from lua script")
	}

	if code == 1 {
		return true, nil
	}
	// We can define specific errors for -1 and -2 if needed, but for now false is enough
	return false, nil
}

func (c *RedisFlashSaleCache) RevertStock(ctx context.Context, flashsaleID, userID uint64, quantity int32) error {
	keyTag := fmt.Sprintf("{flashsale:%d}", flashsaleID)
	// Keys: {flashsale:id}
	// Args: user_id, quantity
	_, err := c.client.Eval(ctx, revertStockScript, []string{keyTag}, userID, quantity).Result()
	return err
}

// Ensure interface implementation
var _ repository.FlashSaleCache = (*RedisFlashSaleCache)(nil)

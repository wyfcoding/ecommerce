package cache

import (
	"context"
	_ "embed" // 用于在编译时将文件内容嵌入到Go程序中。
	"fmt"

	"github.com/redis/go-redis/v9"                                        // 导入Redis客户端库。
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain/repository" // 导入秒杀领域的缓存仓储接口。
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
// 它利用Redis的原子性操作（特别是Lua脚本）来处理高并发场景下的秒杀库存和限购逻辑。
type RedisFlashSaleCache struct {
	client *redis.Client // Redis客户端实例。
}

// NewRedisFlashSaleCache 创建并返回一个新的 RedisFlashSaleCache 实例。
func NewRedisFlashSaleCache(client *redis.Client) *RedisFlashSaleCache {
	return &RedisFlashSaleCache{
		client: client,
	}
}

// SetStock 在Redis中设置指定秒杀活动的库存。
// ctx: 上下文。
// flashsaleID: 秒杀活动ID。
// stock: 要设置的库存数量。
func (c *RedisFlashSaleCache) SetStock(ctx context.Context, flashsaleID uint64, stock int32) error {
	// 构建Redis键，使用哈希标签（{flashsale:ID}）确保相关键位于同一个哈希槽，便于事务或集群操作。
	key := fmt.Sprintf("{flashsale:%d}:stock", flashsaleID)
	// 使用SET命令设置库存，过期时间为0表示永不过期（需要手动管理过期）。
	return c.client.Set(ctx, key, stock, 0).Err()
}

// DeductStock 在Redis中原子性地扣减指定秒杀活动的库存，并检查用户限购。
// 此操作通过执行一个Redis Lua脚本来确保原子性，避免并发问题。
// ctx: 上下文。
// flashsaleID: 秒杀活动ID。
// userID: 用户ID。
// quantity: 待扣减的数量。
// limitPerUser: 每用户限购数量。
// 返回一个布尔值，表示扣减是否成功；以及可能发生的错误。
func (c *RedisFlashSaleCache) DeductStock(ctx context.Context, flashsaleID, userID uint64, quantity, limitPerUser int32) (bool, error) {
	// 构建Redis键的哈希标签，确保所有相关键都在同一个哈希槽。
	keyTag := fmt.Sprintf("{flashsale:%d}", flashsaleID)
	// 执行Lua脚本。
	// KEYS: {flashsale:ID} 包含库存键和用户购买记录键的哈希标签。
	// ARGS: userID, quantity, limitPerUser 作为脚本参数。
	res, err := c.client.Eval(ctx, deductStockScript, []string{keyTag}, userID, quantity, limitPerUser).Result()
	if err != nil {
		return false, err
	}

	// Lua脚本返回1表示成功，-1表示库存不足，-2表示用户购买数量超限。
	code, ok := res.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected return type from lua script")
	}

	if code == 1 {
		return true, nil
	}
	// TODO: 可以根据Lua脚本返回的负数代码，返回更具体的错误类型（例如 ErrNoStock, ErrLimitExceeded）。
	return false, nil
}

// RevertStock 在Redis中原子性地回滚已扣减的库存。
// 此操作也通过执行一个Redis Lua脚本来确保原子性。
// ctx: 上下文。
// flashsaleID: 秒杀活动ID。
// userID: 用户ID。
// quantity: 待回滚的数量。
func (c *RedisFlashSaleCache) RevertStock(ctx context.Context, flashsaleID, userID uint64, quantity int32) error {
	// 构建Redis键的哈希标签。
	keyTag := fmt.Sprintf("{flashsale:%d}", flashsaleID)
	// 执行Lua脚本。
	// KEYS: {flashsale:ID}。
	// ARGS: userID, quantity 作为脚本参数。
	_, err := c.client.Eval(ctx, revertStockScript, []string{keyTag}, userID, quantity).Result()
	return err
}

// Ensure interface implementation 确保 RedisFlashSaleCache 实现了 repository.FlashSaleCache 接口。
var _ repository.FlashSaleCache = (*RedisFlashSaleCache)(nil)

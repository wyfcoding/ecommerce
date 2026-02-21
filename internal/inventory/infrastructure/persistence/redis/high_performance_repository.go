package redis

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
)

// HighPerformanceInventoryRepository 高性能库存仓储
// 专门用于处理高并发秒杀场景，基于 Redis Lua 脚本实现原子扣减
type HighPerformanceInventoryRepository struct {
	client *redis.Client
}

func NewHighPerformanceInventoryRepository(client *redis.Client) *HighPerformanceInventoryRepository {
	return &HighPerformanceInventoryRepository{client: client}
}

// Lua 脚本定义
const (
	// scriptReserveStock: 返回 1=成功, 0=库存不足, -1=重复请求
	scriptReserveStock = `
local stockKey = KEYS[1]
local dedupKey = KEYS[2]
local qty = tonumber(ARGV[1])
local expire = tonumber(ARGV[2])

if redis.call("EXISTS", dedupKey) == 1 then
    return -1
end

local available = tonumber(redis.call("HGET", stockKey, "available") or "0")
if available < qty then
    return 0
end

redis.call("HINCRBY", stockKey, "available", -qty)
redis.call("HINCRBY", stockKey, "reserved", qty)
redis.call("SET", dedupKey, qty, "EX", expire)

return 1
`

	// scriptConfirmStock: 返回 1=成功（实际扣减物理库存）, -1=流水不存在
	scriptConfirmStock = `
local stockKey = KEYS[1]
local dedupKey = KEYS[2]

local qty = redis.call("GET", dedupKey)
if not qty then
    return -1
end

local q = tonumber(qty)
redis.call("HINCRBY", stockKey, "reserved", -q)
redis.call("HINCRBY", stockKey, "total", -q)
redis.call("DEL", dedupKey)

return 1
`

	// scriptRollbackStock: 返回 1=成功, -1=流水不存在
	scriptRollbackStock = `
local stockKey = KEYS[1]
local dedupKey = KEYS[2]

local qty = redis.call("GET", dedupKey)
if not qty then
    return -1
end

local q = tonumber(qty)
redis.call("HINCRBY", stockKey, "available", q)
redis.call("HINCRBY", stockKey, "reserved", -q)
redis.call("DEL", dedupKey)

return 1
`
)

func (r *HighPerformanceInventoryRepository) stockKey(skuID, whID string) string {
	return fmt.Sprintf("stock:%s:%s", whID, skuID)
}

func (r *HighPerformanceInventoryRepository) dedupKey(txID string) string {
	return fmt.Sprintf("stock_tx:%s", txID)
}

// Reserve 高性能库存预占
func (r *HighPerformanceInventoryRepository) Reserve(ctx context.Context, req domain.DeductRequest) (bool, error) {
	keys := []string{r.stockKey(req.SKUID, req.WarehouseID), r.dedupKey(req.TxID)}
	// 预占保留时间：30分钟
	expiry := 1800

	res, err := r.client.Eval(ctx, scriptReserveStock, keys, req.Quantity, expiry).Int()
	if err != nil {
		return false, err
	}

	switch res {
	case -1:
		return true, nil // 幂等成功
	case 0:
		return false, errors.New("insufficient stock")
	case 1:
		return true, nil
	default:
		return false, errors.New("unexpected result from redis script")
	}
}

// ConfirmWithContext 确认扣减
func (r *HighPerformanceInventoryRepository) ConfirmWithContext(ctx context.Context, txID, skuID, whID string) error {
	keys := []string{r.stockKey(skuID, whID), r.dedupKey(txID)}
	res, err := r.client.Eval(ctx, scriptConfirmStock, keys).Int()
	if err != nil {
		return err
	}
	if res == -1 {
		return errors.New("transaction not found or expired")
	}
	return nil
}

// RollbackWithContext 回滚预占
func (r *HighPerformanceInventoryRepository) RollbackWithContext(ctx context.Context, txID, skuID, whID string) error {
	keys := []string{r.stockKey(skuID, whID), r.dedupKey(txID)}
	res, err := r.client.Eval(ctx, scriptRollbackStock, keys).Int()
	if err != nil {
		return err
	}
	if res == -1 {
		// 可能是已过期或已回滚，视作成功以保证幂等
		return nil
	}
	return nil
}

// Get 获取库存信息
func (r *HighPerformanceInventoryRepository) Get(ctx context.Context, skuID, warehouseID string) (*domain.InventoryItem, error) {
	key := r.stockKey(skuID, warehouseID)
	data, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("stock not found")
	}

	// 解析 Redis 哈希数据到 InventoryItem
	// 这里需要根据实际数据结构进行解析
	return &domain.InventoryItem{
		SKUID:       skuID,
		WarehouseID: warehouseID,
	}, nil
}

// 实现 InventoryRepository 接口的其他方法
func (r *HighPerformanceInventoryRepository) Confirm(ctx context.Context, txID string) error {
	return errors.New("not implemented: need proper context mapping")
}

func (r *HighPerformanceInventoryRepository) Rollback(ctx context.Context, txID string) error {
	return errors.New("not implemented: need proper context mapping")
}

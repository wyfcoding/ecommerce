package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Lua Script: Atomic Stock Deduction & User Limit Check
// KEYS[1]: stock_key (e.g., "seckill:stock:{id}")
// KEYS[2]: user_history_key (e.g., "seckill:user:{id}")
// ARGV[1]: user_id
// ARGV[2]: limit_per_user
// ARGV[3]: buy_quantity
const SeckillLuaScript = `
	local stock_key = KEYS[1]
	local user_key = KEYS[2]
	local user_id = ARGV[1]
	local limit = tonumber(ARGV[2])
	local quantity = tonumber(ARGV[3])

	-- 1. Check Stock
	local stock = tonumber(redis.call('get', stock_key) or "0")
	if stock < quantity then
		return -1 -- Sold Out
	end

	-- 2. Check User Limit (using Hash: {user_id: bought_count})
	-- Note: In high concurrency, we might just use a Set if limit is 1, or Hash if limit > 1
	local bought = tonumber(redis.call('hget', user_key, user_id) or "0")
	if limit > 0 and (bought + quantity) > limit then
		return -2 -- Limit Exceeded
	end

	-- 3. Deduct Stock & Record User
	redis.call('decrby', stock_key, quantity)
	redis.call('hincrby', user_key, user_id, quantity)
	
	return 1 -- Success
`

var (
	ErrSeckillSoldOut = errors.New("seckill: sold out")
	ErrSeckillLimit   = errors.New("seckill: purchase limit exceeded")
	ErrSeckillSystem  = errors.New("seckill: system error")
)

// SeckillEngine 核心秒杀引擎 (Redis + Lua)
type SeckillEngine struct {
	redisClient *redis.Client
	scriptSha   string
}

func NewSeckillEngine(rdb *redis.Client) *SeckillEngine {
	// Pre-load script
	return &SeckillEngine{
		redisClient: rdb,
	}
}

// LoadScript ensures the Lua script is loaded in Redis
func (e *SeckillEngine) LoadScript(ctx context.Context) error {
	sha, err := e.redisClient.ScriptLoad(ctx, SeckillLuaScript).Result()
	if err != nil {
		return fmt.Errorf("failed to load lua script: %w", err)
	}
	e.scriptSha = sha
	return nil
}

// Execute 尝试执行秒杀扣减
// 返回 true 表示抢购成功（进入异步下单流程），返回 error 表示失败
func (e *SeckillEngine) Execute(ctx context.Context, flashsaleID uint64, userID uint64, quantity int32, limit int32) (bool, error) {
	if e.scriptSha == "" {
		if err := e.LoadScript(ctx); err != nil {
			return false, err
		}
	}

	stockKey := fmt.Sprintf("seckill:stock:%d", flashsaleID)
	userKey := fmt.Sprintf("seckill:users:%d", flashsaleID)

	// Execute Lua
	res, err := e.redisClient.EvalSha(ctx, e.scriptSha, []string{stockKey, userKey}, userID, limit, quantity).Int()
	if err != nil {
		// Try to reload script if NOSCRIPT error
		if isNoScriptErr(err) {
			_ = e.LoadScript(ctx)
			res, err = e.redisClient.EvalSha(ctx, e.scriptSha, []string{stockKey, userKey}, userID, limit, quantity).Int()
		}
		if err != nil {
			return false, fmt.Errorf("redis execution failed: %w", err)
		}
	}

	switch res {
	case 1:
		return true, nil
	case -1:
		return false, ErrSeckillSoldOut
	case -2:
		return false, ErrSeckillLimit
	default:
		return false, ErrSeckillSystem
	}
}

// PreheatStock 预热库存到 Redis
func (e *SeckillEngine) PreheatStock(ctx context.Context, flashsaleID uint64, stock int32) error {
	stockKey := fmt.Sprintf("seckill:stock:%d", flashsaleID)
	return e.redisClient.Set(ctx, stockKey, stock, 24*time.Hour).Err()
}

func isNoScriptErr(err error) bool {
	return err != nil && err.Error() == "NOSCRIPT No matching script. Please use EVAL."
}

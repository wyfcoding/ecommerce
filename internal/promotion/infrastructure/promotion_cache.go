// 变更说明：
// 促销额度的高并发 Redis Lua 拦截器实现。
package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/promotion/domain"
)

// 针对秒杀或限量促销的核心 Lua 防超发防护网
// KEYS[1] = global_quota_key, KEYS[2] = user_quota_key
// ARGV[1] = global_limit, ARGV[2] = user_limit, ARGV[3] = user_id
const promoteQuotaLua = `
	local g_limit = tonumber(ARGV[1])
	local u_limit = tonumber(ARGV[2])
	local u_id = ARGV[3]

	-- 用户维度的限制防控
	if u_limit > 0 then
		local u_used = tonumber(redis.call('HGET', KEYS[2], u_id) or "0")
		if u_used >= u_limit then
			return -2 -- 用户额度耗尽
		end
	end

	-- 全局维度的限制防控
	if g_limit > 0 then
		local g_used = tonumber(redis.call('GET', KEYS[1]) or "0")
		if g_used >= g_limit then
			return -1 -- 全局大盘名额耗尽
		end
	end

	-- 执行真实扣减
	if u_limit > 0 then redis.call('HINCRBY', KEYS[2], u_id, 1) end
	if g_limit > 0 then redis.call('INCR', KEYS[1]) end

	return 1 -- 成功
`

type PromotionCacheImpl struct {
	rdb       redis.UniversalClient
	scriptSha string
}

func NewPromotionCache(rdb redis.UniversalClient) (domain.PromotionCache, error) {
	pc := &PromotionCacheImpl{rdb: rdb}
	// 启动时预热 Lua 脚本进内存
	if err := pc.loadScript(context.Background()); err != nil {
		return nil, err
	}
	return pc, nil
}

func (c *PromotionCacheImpl) loadScript(ctx context.Context) error {
	sha, err := c.rdb.ScriptLoad(ctx, promoteQuotaLua).Result()
	if err != nil {
		return fmt.Errorf("failed to load lua script: %v", err)
	}
	c.scriptSha = sha
	return nil
}

func (c *PromotionCacheImpl) DeductQuota(ctx context.Context, promotionID, userID uint64, globalLimit int64, userLimit int32) (int, error) {
	gKey := fmt.Sprintf("promo:quota:global:%d", promotionID)
	uKey := fmt.Sprintf("promo:quota:user:%d", promotionID)

	res, err := c.rdb.EvalSha(ctx, c.scriptSha, []string{gKey, uKey}, globalLimit, userLimit, userID).Int()
	if err != nil {
		// 容错：当连接的是新Redis集群，脚本不存在时，触发 Reload
		if err.Error() == "NOSCRIPT No matching script. Please use EVAL." {
			_ = c.loadScript(ctx)
			res, err = c.rdb.EvalSha(ctx, c.scriptSha, []string{gKey, uKey}, globalLimit, userLimit, userID).Int()
		}
		if err != nil {
			return 0, fmt.Errorf("lua evalsha failed: %w", err)
		}
	}
	return res, nil
}

func (c *PromotionCacheImpl) RollbackQuota(ctx context.Context, promotionID, userID uint64) error {
	// 用于回退扣减，取消订单时调用 (使用 Pipeline 降低网络 RTT)
	pipe := c.rdb.Pipeline()
	pipe.Decr(ctx, fmt.Sprintf("promo:quota:global:%d", promotionID))
	pipe.HIncrBy(ctx, fmt.Sprintf("promo:quota:user:%d", promotionID), fmt.Sprintf("%d", userID), -1)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *PromotionCacheImpl) GetActiveByProduct(ctx context.Context, productID uint64) ([]*domain.Promotion, error) {
	key := fmt.Sprintf("promo:cache:product:%d", productID)
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // miss
	} else if err != nil {
		return nil, err
	}

	var promos []*domain.Promotion
	if err := json.Unmarshal([]byte(val), &promos); err != nil {
		return nil, err
	}
	return promos, nil
}

func (c *PromotionCacheImpl) SetProductPromotions(ctx context.Context, productID uint64, promos []*domain.Promotion) error {
	key := fmt.Sprintf("promo:cache:product:%d", productID)
	bs, _ := json.Marshal(promos)
	// 过期时间加入随机抖动，防止雪崩
	ttl := 10*time.Minute + time.Duration(time.Now().UnixNano()%300)*time.Second
	return c.rdb.Set(ctx, key, bs, ttl).Err()
}

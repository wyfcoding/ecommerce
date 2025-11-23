package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

// Limiter 限流器接口
type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

// LocalLimiter 本地限流器（基于令牌桶）
type LocalLimiter struct {
	limiter *rate.Limiter
}

// NewLocalLimiter 创建本地限流器
// r: 每秒生成的令牌数
// b: 令牌桶容量
func NewLocalLimiter(r rate.Limit, b int) *LocalLimiter {
	return &LocalLimiter{
		limiter: rate.NewLimiter(r, b),
	}
}

// Allow 检查是否允许请求
func (l *LocalLimiter) Allow(ctx context.Context, key string) (bool, error) {
	return l.limiter.Allow(), nil
}

// RedisLimiter 分布式限流器（基于Redis）
type RedisLimiter struct {
	client *redis.Client
	limit  int           // 时间窗口内的最大请求数
	window time.Duration // 时间窗口
}

// NewRedisLimiter 创建Redis限流器
func NewRedisLimiter(client *redis.Client, limit int, window time.Duration) *RedisLimiter {
	return &RedisLimiter{
		client: client,
		limit:  limit,
		window: window,
	}
}

// Allow 检查是否允许请求（滑动窗口算法）
func (l *RedisLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixNano()
	windowStart := now - l.window.Nanoseconds()

	pipe := l.client.Pipeline()

	// 删除窗口外的记录
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// 统计当前窗口内的请求数
	pipe.ZCard(ctx, key)

	// 添加当前请求
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now),
		Member: now,
	})

	// 设置过期时间
	pipe.Expire(ctx, key, l.window)

	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	// 获取当前窗口内的请求数
	count := cmds[1].(*redis.IntCmd).Val()

	return count < int64(l.limit), nil
}

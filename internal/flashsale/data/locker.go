package data

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// redisLocker implements biz.DistributedLocker using Redis.
type redisLocker struct {
	redisClient *redis.Client
}

// AcquireLock attempts to acquire a distributed lock.
// It uses Redis's SET NX EX command.
func (l *redisLocker) AcquireLock(ctx context.Context, key string, expiry time.Duration) (bool, error) {
	cmd := l.redisClient.SetNX(ctx, key, "locked", expiry)
	return cmd.Result()
}

// ReleaseLock releases a distributed lock.
// It uses Redis's DEL command.
func (l *redisLocker) ReleaseLock(ctx context.Context, key string) (bool, error) {
	cmd := l.redisClient.Del(ctx, key)
	return cmd.Result()
}

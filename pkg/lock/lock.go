package lock

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrLockFailed   = errors.New("failed to acquire lock")
	ErrUnlockFailed = errors.New("failed to release lock")
)

// DistributedLock 分布式锁接口
type DistributedLock interface {
	Lock(ctx context.Context, key string, ttl time.Duration) (string, error)
	Unlock(ctx context.Context, key string, token string) error
}

// RedisLock Redis分布式锁
type RedisLock struct {
	client *redis.Client
}

// NewRedisLock 创建Redis分布式锁
func NewRedisLock(client *redis.Client) *RedisLock {
	return &RedisLock{client: client}
}

// Lock 获取锁
func (l *RedisLock) Lock(ctx context.Context, key string, ttl time.Duration) (string, error) {
	token := generateToken()
	ok, err := l.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrLockFailed
	}
	return token, nil
}

// Unlock 释放锁
func (l *RedisLock) Unlock(ctx context.Context, key string, token string) error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	result, err := l.client.Eval(ctx, script, []string{key}, token).Result()
	if err != nil {
		return err
	}
	if result.(int64) == 0 {
		return ErrUnlockFailed
	}
	return nil
}

// generateToken 生成唯一token
func generateToken() string {
	return time.Now().Format("20060102150405.000000")
}

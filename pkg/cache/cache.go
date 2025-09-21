package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Cache 定义了通用的缓存接口。
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
	// ... 其他常用缓存操作，如 Incr, Decr, HSet, HGet, HMGet, HDel, Expire, TTL 等
}

// RedisCache 是 Cache 接口基于 Redis 的实现。
type RedisCache struct {
	client *redis.Client
}

// RedisConfig 结构体用于 Redis 客户端配置。
type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewRedisCache 创建一个新的 RedisCache 实例。
func NewRedisCache(conf *RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         conf.Addr,
		Password:     conf.Password,
		DB:           conf.DB,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
	})

	// Ping Redis 以检查连接是否正常
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

// Set 将键值对存储到缓存中，并设置过期时间。
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return rc.client.Set(ctx, key, value, expiration).Err()
}

// Get 从缓存中获取指定键的值。
func (rc *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return rc.client.Get(ctx, key).Result()
}

// Del 从缓存中删除指定键。
func (rc *RedisCache) Del(ctx context.Context, key string) error {
	return rc.client.Del(ctx, key).Err()
}

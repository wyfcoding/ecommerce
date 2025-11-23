package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"ecommerce/pkg/config"

	"github.com/redis/go-redis/v9"
)

// Cache defines the cache interface.
type Cache interface {
	Get(ctx context.Context, key string, value interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)
	Close() error
}

// RedisCache implements Cache using Redis.
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewRedisCache creates a new RedisCache instance.
func NewRedisCache(cfg config.RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &RedisCache{
		client: client,
		prefix: "", // Prefix can be handled by caller or config if needed, but usually key namespacing is better
	}, nil
}

// WithPrefix returns a new RedisCache with a key prefix.
func (c *RedisCache) WithPrefix(prefix string) *RedisCache {
	return &RedisCache{
		client: c.client,
		prefix: prefix,
	}
}

// buildKey builds the key with prefix.
func (c *RedisCache) buildKey(key string) string {
	if c.prefix == "" {
		return key
	}
	return c.prefix + ":" + key
}

// Get retrieves a value from cache.
func (c *RedisCache) Get(ctx context.Context, key string, value interface{}) error {
	fullKey := c.buildKey(key)
	data, err := c.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("cache miss: %s", key)
		}
		return fmt.Errorf("cache get error: %w", err)
	}
	return json.Unmarshal(data, value)
}

// Set sets a value in cache.
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	fullKey := c.buildKey(key)
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	return c.client.Set(ctx, fullKey, data, expiration).Err()
}

// Delete deletes keys from cache.
func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = c.buildKey(key)
	}
	return c.client.Del(ctx, fullKeys...).Err()
}

// Exists checks if a key exists.
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.buildKey(key)
	n, err := c.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// Close closes the Redis client.
func (c *RedisCache) Close() error {
	slog.Info("closing redis cache connection...")
	return c.client.Close()
}

// GetClient returns the underlying Redis client.
func (c *RedisCache) GetClient() *redis.Client {
	return c.client
}

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/pkg/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
)

var (
	cacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "The total number of cache hits",
		},
		[]string{"prefix"},
	)
	cacheMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "The total number of cache misses",
		},
		[]string{"prefix"},
	)
	cacheDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cache_operation_duration_seconds",
			Help:    "The duration of cache operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"prefix", "operation"},
	)
)

func init() {
	prometheus.MustRegister(cacheHits, cacheMisses, cacheDuration)
}

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
	cb     *gobreaker.CircuitBreaker
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

	// Initialize Circuit Breaker
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "redis-cache",
		MaxRequests: 0,
		Interval:    0,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.6
		},
	})

	return &RedisCache{
		client: client,
		prefix: "",
		cb:     cb,
	}, nil
}

// WithPrefix returns a new RedisCache with a key prefix.
func (c *RedisCache) WithPrefix(prefix string) *RedisCache {
	return &RedisCache{
		client: c.client,
		prefix: prefix,
		cb:     c.cb,
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
	start := time.Now()
	defer func() {
		cacheDuration.WithLabelValues(c.prefix, "get").Observe(time.Since(start).Seconds())
	}()

	fullKey := c.buildKey(key)

	_, err := c.cb.Execute(func() (interface{}, error) {
		data, err := c.client.Get(ctx, fullKey).Bytes()
		if err != nil {
			if err == redis.Nil {
				cacheMisses.WithLabelValues(c.prefix).Inc()
				return nil, fmt.Errorf("cache miss: %s", key)
			}
			return nil, err
		}
		cacheHits.WithLabelValues(c.prefix).Inc()
		return data, json.Unmarshal(data, value)
	})

	return err
}

// Set sets a value in cache.
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	start := time.Now()
	defer func() {
		cacheDuration.WithLabelValues(c.prefix, "set").Observe(time.Since(start).Seconds())
	}()

	fullKey := c.buildKey(key)
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	_, err = c.cb.Execute(func() (interface{}, error) {
		return nil, c.client.Set(ctx, fullKey, data, expiration).Err()
	})

	return err
}

// Delete deletes keys from cache.
func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	start := time.Now()
	defer func() {
		cacheDuration.WithLabelValues(c.prefix, "delete").Observe(time.Since(start).Seconds())
	}()

	if len(keys) == 0 {
		return nil
	}
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = c.buildKey(key)
	}

	_, err := c.cb.Execute(func() (interface{}, error) {
		return nil, c.client.Del(ctx, fullKeys...).Err()
	})

	return err
}

// Exists checks if a key exists.
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	start := time.Now()
	defer func() {
		cacheDuration.WithLabelValues(c.prefix, "exists").Observe(time.Since(start).Seconds())
	}()

	fullKey := c.buildKey(key)

	result, err := c.cb.Execute(func() (interface{}, error) {
		n, err := c.client.Exists(ctx, fullKey).Result()
		if err != nil {
			return false, err
		}
		return n > 0, nil
	})

	if err != nil {
		return false, err
	}
	return result.(bool), nil
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

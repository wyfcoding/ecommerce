package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/allegro/bigcache/v3"
)

// BigCache implements Cache using BigCache (In-Memory).
type BigCache struct {
	cache *bigcache.BigCache
}

// NewBigCache creates a new BigCache instance.
func NewBigCache(ttl time.Duration, maxMB int) (*BigCache, error) {
	config := bigcache.DefaultConfig(ttl)
	config.HardMaxCacheSize = maxMB
	config.CleanWindow = 5 * time.Minute

	cache, err := bigcache.New(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to init bigcache: %w", err)
	}

	return &BigCache{cache: cache}, nil
}

func (c *BigCache) Get(ctx context.Context, key string, value interface{}) error {
	data, err := c.cache.Get(key)
	if err != nil {
		if err == bigcache.ErrEntryNotFound {
			return fmt.Errorf("cache miss: %s", key)
		}
		return err
	}
	return json.Unmarshal(data, value)
}

func (c *BigCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// BigCache doesn't support per-key TTL in Set, it uses global TTL.
	// We ignore expiration here or we could encode it in value.
	// For simplicity, we rely on global TTL for L1.
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.cache.Set(key, data)
}

func (c *BigCache) Delete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		if err := c.cache.Delete(key); err != nil {
			// Ignore not found error
			if err != bigcache.ErrEntryNotFound {
				return err
			}
		}
	}
	return nil
}

func (c *BigCache) Exists(ctx context.Context, key string) (bool, error) {
	_, err := c.cache.Get(key)
	if err == nil {
		return true, nil
	}
	if err == bigcache.ErrEntryNotFound {
		return false, nil
	}
	return false, err
}

func (c *BigCache) Close() error {
	return c.cache.Close()
}

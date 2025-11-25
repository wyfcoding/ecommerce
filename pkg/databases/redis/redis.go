package redis

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/pkg/config"
	"github.com/wyfcoding/ecommerce/pkg/logging"

	"github.com/redis/go-redis/v9"
)

// The original Config struct is likely replaced by config.RedisConfig
// and the NewRedis function directly uses config.RedisConfig.
// If this local Config struct is no longer used, it can be removed.
// For now, I'll keep it but it seems redundant with the new NewRedis signature.
type Config struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
}

// NewRedis creates a new Redis client using the provided configuration and logger.
// It replaces the previous NewRedisClient which used a local Config struct
// and returned a cleanup function.
func NewRedis(cfg *config.RedisConfig, logger *logging.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		// PoolSize, MinIdleConns, ReadTimeout, WriteTimeout can be added here
		// if they are part of config.RedisConfig and needed for direct client creation.
		// For now, using the basic fields from the provided snippet.
		ReadTimeout:  3 * time.Second, // Default
		WriteTimeout: 3 * time.Second, // Default
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		return nil, err
	}

	logger.Info("Successfully connected to Redis", "addr", client.Options().Addr)

	return client, nil
}

// The original NewRedisClient function is removed as per the instruction
// which provides a new NewRedis function.
// If the cache.NewRedisCache path is still needed, this function would need to be adapted
// or a separate function created. The instruction implies a direct redis.Client creation.

// The following is the original NewRedisClient function, commented out as it's replaced.
/*
func NewRedisClient(cfg *Config) (*redis.Client, func(), error) {
	redisConfig := config.RedisConfig{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		ReadTimeout:  3 * time.Second, // Default
		WriteTimeout: 3 * time.Second, // Default
	}

	c, err := cache.NewRedisCache(redisConfig)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		c.Close()
	}

	return c.GetClient(), cleanup, nil
}
*/

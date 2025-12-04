package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/pkg/config"
	"github.com/wyfcoding/ecommerce/pkg/logging"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a new Redis client using the provided configuration.
// It returns a *redis.Client instance, a cleanup function, and an error if connection fails.
func NewRedisClient(cfg *config.RedisConfig, logger *logging.Logger) (*redis.Client, func(), error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	// 创建一个带超时机制的上下文，用于Ping操作。
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // 确保上下文在函数退出时被取消。

	// 尝试Ping Redis服务器以验证连接的可用性。
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Successfully connected to Redis", "addr", client.Options().Addr)

	// 定义一个清理函数，用于在应用程序关闭时优雅地关闭Redis客户端。
	cleanup := func() {
		if err := client.Close(); err != nil {
			logger.Error("failed to close Redis client", "error", err)
		}
	}

	return client, cleanup, nil
}

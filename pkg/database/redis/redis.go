package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// Config 结构体用于 Redis 客户端配置。
type Config struct {
	Addr         string        `toml:"addr"`
	Password     string        `toml:"password"`
	DB           int           `toml:"db"`
	ReadTimeout  time.Duration `toml:"read_timeout"`
	WriteTimeout time.Duration `toml:"write_timeout"`
	PoolSize     int           `toml:"pool_size"`
	MinIdleConns int           `toml:"min_idle_conns"`
}

// NewRedisClient 创建一个新的 Redis 客户端实例。
func NewRedisClient(conf *Config) (*redis.Client, func(), error) {
	client := redis.NewClient(&redis.Options{
		Addr:         conf.Addr,
		Password:     conf.Password,
		DB:           conf.DB,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
		PoolSize:     conf.PoolSize,
		MinIdleConns: conf.MinIdleConns,
	})

	// Ping Redis 以检查连接是否正常
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	cleanup := func() {
		zap.S().Info("closing redis connection...")
		if err := client.Close(); err != nil {
			zap.S().Errorf("failed to close redis connection: %v", err)
		}
	}

	return client, cleanup, nil
}

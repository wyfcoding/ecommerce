package repository

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type Data struct {
	rdb *redis.Client
	log *zap.SugaredLogger // 添加日志器
}

// NewData 是 Data 结构体的构造函数
// 它接收 Redis 配置，初始化 Redis 客户端，并返回一个包含连接的 Data 实例
func NewData(conf *RedisConfig, logger *zap.SugaredLogger) (*Data, func(), error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         conf.Addr,
		Password:     conf.Password,
		DB:           conf.DB,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
	})

	// Ping Redis 以检查连接是否正常
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		logger.Errorf("failed to connect to redis: %v", err) // 使用注入的 logger
		return nil, nil, err
	}

	// 定义一个清理函数，用于在服务关闭时关闭 Redis 连接
	cleanup := func() {
		logger.Info("closing redis connection...") // 使用注入的 logger
		if err := rdb.Close(); err != nil {
			logger.Errorf("failed to close redis connection: %v", err) // 使用注入的 logger
		}
	}

	return &Data{rdb: rdb, log: logger}, cleanup, nil // 初始化 log 字段
}

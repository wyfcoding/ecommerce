package data

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Data 结构体持有数据库连接和 Redis 客户端。
type Data struct {
	db  *gorm.DB
	rdb *redis.Client
	log *zap.SugaredLogger // 添加日志器
}

// RedisConfig 结构体用于映射 Redis 配置
type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewData 是 Data 结构体的构造函数。
func NewData(db *gorm.DB, redisConf *RedisConfig, logger *zap.SugaredLogger) (*Data, func(), error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         redisConf.Addr,
		Password:     redisConf.Password,
		DB:           redisConf.DB,
		ReadTimeout:  redisConf.ReadTimeout,
		WriteTimeout: redisConf.WriteTimeout,
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

	return &Data{db: db, rdb: rdb, log: logger}, cleanup, nil // 初始化 log 字段
}

// InTx 在一个数据库事务中执行函数。
func (d *Data) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return d.db.WithContext(ctx).Transaction(fn)
}

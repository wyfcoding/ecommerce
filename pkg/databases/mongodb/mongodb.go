package mongodb

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Config 结构体用于 MongoDB 数据库配置。
type Config struct {
	URI            string        `toml:"uri"`
	Database       string        `toml:"database"`
	ConnectTimeout time.Duration `toml:"connect_timeout"`
	MinPoolSize    uint64        `toml:"min_pool_size"`
	MaxPoolSize    uint64        `toml:"max_pool_size"`
}

// NewMongoClient 创建一个新的 MongoDB 客户端连接。
func NewMongoClient(conf *Config) (*mongo.Client, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), conf.ConnectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.URI).SetMinPoolSize(conf.MinPoolSize).SetMaxPoolSize(conf.MaxPoolSize))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	// Ping the primary to verify connection
	ctx, cancel = context.WithTimeout(context.Background(), conf.ConnectTimeout)
	defer cancel()
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			slog.Error("failed to disconnect from mongodb", "error", err)
		}
	}

	return client, cleanup, nil
}

package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.uber.org/zap"
)

// Config 结构体用于 ClickHouse 数据库配置。
type Config struct {
	DSN             string        `toml:"dsn"`
	Database        string        `toml:"database"`
	Username        string        `toml:"username"`
	Password        string        `toml:"password"`
	DialTimeout     time.Duration `toml:"dial_timeout"`
	MaxOpenConns    int           `toml:"max_open_conns"`
	MaxIdleConns    int           `toml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `toml:"conn_max_lifetime"`
}

// NewClickHouseClient 创建一个新的 ClickHouse 客户端连接。
func NewClickHouseClient(conf *Config) (clickhouse.Conn, func(), error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{conf.DSN},
		Auth: clickhouse.Auth{
			Database: conf.Database,
			Username: conf.Username,
			Password: conf.Password,
		},
		DialTimeout:     conf.DialTimeout,
		MaxOpenConns:    conf.MaxOpenConns,
		MaxIdleConns:    conf.MaxIdleConns,
		ConnMaxLifetime: conf.ConnMaxLifetime,
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		// Debug: true, // Uncomment for debugging
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	// Ping the database to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), conf.DialTimeout)
	defer cancel()
	if err := conn.Ping(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	cleanup := func() {
		if conn != nil {
			zap.S().Info("closing clickhouse connection...")
			if err := conn.Close(); err != nil {
				zap.S().Errorf("failed to close clickhouse connection: %v", err)
			}
		}
	}

	return conn, cleanup, nil
}

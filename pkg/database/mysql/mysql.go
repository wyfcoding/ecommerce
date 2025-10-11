package mysql

import (
	"fmt"
	"time"

	"ecommerce/pkg/gormlogger"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config 结构体用于 MySQL 数据库配置。
type Config struct {
	DSN             string        `toml:"dsn"`
	MaxIdleConns    int           `toml:"max_idle_conns"`
	MaxOpenConns    int           `toml:"max_open_conns"`
	ConnMaxLifetime time.Duration `toml:"conn_max_lifetime"`
	LogLevel        logger.LogLevel `toml:"log_level"`
	SlowThreshold   time.Duration `toml:"slow_threshold"`
}

// NewGORMDB 创建一个新的 GORM 数据库连接实例。
func NewGORMDB(conf *Config, zapLogger *zap.Logger) (*gorm.DB, func(), error) {
	// 使用 GORM 连接到 MySQL 数据库
	db, err := gorm.Open(mysql.Open(conf.DSN), &gorm.Config{
		Logger: gormlogger.New(zapLogger, conf.LogLevel, conf.SlowThreshold),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 获取底层数据库连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(conf.MaxIdleConns)
	sqlDB.SetMaxOpenConns(conf.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(conf.ConnMaxLifetime)

	// 定义一个清理函数，用于在服务关闭时关闭数据库连接
	cleanup := func() {
		if sqlDB != nil {
			zap.S().Info("closing database connection...")
			if err := sqlDB.Close(); err != nil {
				zap.S().Errorf("failed to close database connection: %v", err)
			}
		}
	}

	return db, cleanup, nil
}

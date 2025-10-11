package main

import (
	"fmt"
	"time"

	"ecommerce/api/settlement/v1"
	"ecommerce/internal/settlement/repository"
	"ecommerce/internal/settlement/service"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	"ecommerce/pkg/database/redis"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/snowflake"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Config is the service-specific configuration structure.
type Config struct {
	configpkg.Config
	Data struct {
		configpkg.DataConfig
		Database struct {
			configpkg.DatabaseConfig
			LogLevel      gormlogger.LogLevel `toml:"log_level"`
			SlowThreshold time.Duration     `toml:"slow_threshold"`
		} `toml:"database"`
		Redis redis.Config `toml:"redis"`
	} `toml:"data"`
}

func main() {
	app.NewBuilder("settlement").
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithMetrics("9094").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	v1.RegisterSettlementServiceServer(s, srv.(v1.SettlementServiceServer))
}

func registerGin(e *gin.Engine, srv interface{}) {
	// Placeholder for Gin handlers
}

// NOTE: This service appears to be broken. The 'biz' directory is missing.
// The following initService function is a reconstruction based on the available files.
func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)

	snowflakeNode, cleanupSnowflake, err := snowflake.NewSnowflakeNode(&config.Snowflake)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to new snowflake node: %w", err)
	}

	db, err := gorm.Open(mysql.Open(config.Data.Database.DSN), &gorm.Config{
		Logger: gormlogger.New(zap.L(), config.Data.Database.LogLevel, config.Data.Database.SlowThreshold),
	})
	if err != nil {
		cleanupSnowflake()
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		cleanupSnowflake()
		return nil, nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	data, cleanupData, err := repository.NewData(db)
	if err != nil {
		cleanupSnowflake()
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to create data struct: %w", err)
	}

	redisClient, cleanupRedis, err := redis.NewRedisClient(&config.Data.Redis)
	if err != nil {
		cleanupData()
		cleanupSnowflake()
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to new redis client: %w", err)
	}

	settlementRepo := repository.NewSettlementRepo(data, redisClient, snowflakeNode) // Assuming this is the correct signature
	settlementUsecase := service.NewSettlementUsecase(settlementRepo)      // Assuming this exists in the service package
	settlementService := service.NewSettlementService(settlementUsecase)

	cleanup := func() {
		cleanupRedis()
		cleanupData()
		sqlDB.Close()
		cleanupSnowflake()
	}

	return settlementService, cleanup, nil
}
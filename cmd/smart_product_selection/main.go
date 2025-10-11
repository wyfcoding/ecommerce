package main

import (
	"fmt"
	"time"

	"ecommerce/api/smart_product_selection/v1"
	"ecommerce/internal/smart_product_selection/biz"
	"ecommerce/internal/smart_product_selection/data"
	"ecommerce/internal/smart_product_selection/handler"
	"ecommerce/internal/smart_product_selection/service"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	"ecommerce/pkg/database/redis"
	"ecommerce/pkg/metrics"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"go.uber.org/zap"
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
	app.NewBuilder("smart_product_selection").
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithMetrics("9098").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	v1.RegisterSmartProductSelectionServiceServer(s, srv.(v1.SmartProductSelectionServiceServer))
}

func registerGin(e *gin.Engine, srv interface{}) {
	// Placeholder for Gin handlers
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)

	dataInstance, cleanupData, err := data.NewData(config.Data.Database.DSN, zap.L(), config.Data.Database.LogLevel, config.Data.Database.SlowThreshold)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to new data: %w", err)
	}

	redisClient, cleanupRedis, err := redis.NewRedisClient(&config.Data.Redis)
	if err != nil {
		cleanupData()
		return nil, nil, fmt.Errorf("failed to new redis client: %w", err)
	}

	smartProductSelectionRepo := data.NewSmartProductSelectionRepo(dataInstance, redisClient)
	smartProductSelectionUsecase := biz.NewSmartProductSelectionUsecase(smartProductSelectionRepo)
	smartProductSelectionService := service.NewSmartProductSelectionService(smartProductSelectionUsecase)

	cleanup := func() {
		cleanupRedis()
		cleanupData()
	}

	return smartProductSelectionService, cleanup, nil
}
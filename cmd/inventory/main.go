package main

import (
	"fmt"
	"time"

	"ecommerce/internal/inventory/repository"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	mysqlpkg "ecommerce/pkg/database/mysql"
	redispkg "ecommerce/pkg/database/redis"
	"ecommerce/pkg/lock"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/tracing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Config struct {
	configpkg.Config
}

const serviceName = "inventory-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9096").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	zap.S().Info("gRPC server registered for inventory service")
}

func registerGin(e *gin.Engine, srv interface{}) {
	api := e.Group("/api/v1/inventory")
	{
		api.GET("/:sku_id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "get inventory"})
		})
		api.POST("/deduct", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "deduct stock"})
		})
		api.POST("/lock", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "lock stock"})
		})
	}
	zap.S().Info("HTTP routes registered for inventory service")
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)
	cleanupTracer := func() {}

	mysqlConfig := &mysqlpkg.Config{
		DSN:             config.Data.Database.DSN,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		LogLevel:        4,
		SlowThreshold:   200 * time.Millisecond,
	}
	db, cleanupDB, err := mysqlpkg.NewGORMDB(mysqlConfig, zap.L())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	redisConfig := &redispkg.Config{
		Addr:         config.Data.Redis.Addr,
		Password:     config.Data.Redis.Password,
		DB:           config.Data.Redis.DB,
		PoolSize:     10,
		MinIdleConns: 5,
	}
	redisClient, cleanupRedis, err := redispkg.NewRedisClient(redisConfig)
	if err != nil {
		cleanupDB()
		return nil, nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	distributedLock := lock.NewRedisLock(redisClient)
	inventoryRepo := repository.NewInventoryRepository(db, redisClient, distributedLock)
	_ = inventoryRepo

	cleanup := func() {
		zap.S().Info("cleaning up inventory service resources...")
		cleanupRedis()
		cleanupDB()
		cleanupTracer()
	}

	return nil, cleanup, nil
}

package main

import (
	"fmt"
	"time"

	"ecommerce/internal/cart/repository"
	"ecommerce/internal/cart/service"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	mysqlpkg "ecommerce/pkg/database/mysql"
	redispkg "ecommerce/pkg/database/redis"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/tracing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Config struct {
	configpkg.Config
}

const serviceName = "cart-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9094").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	// TODO: 注册gRPC服务
	// v1.RegisterCartServer(s, srv.(*service.CartService))
	zap.S().Info("gRPC server registered")
}

func registerGin(e *gin.Engine, srv interface{}) {
	// TODO: 注册HTTP路由
	zap.S().Info("HTTP routes registered")
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)

	cleanupTracer := func() {}

	// 初始化MySQL
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

	// 初始化Redis
	redisConfig := &redispkg.Config{
		Addr:         config.Data.Redis.Addr,
		Password:     config.Data.Redis.Password,
		DB:           config.Data.Redis.DB,
		ReadTimeout:  config.Data.Redis.ReadTimeout,
		WriteTimeout: config.Data.Redis.WriteTimeout,
		PoolSize:     config.Data.Redis.PoolSize,
		MinIdleConns: config.Data.Redis.MinIdleConns,
	}
	redisClient, cleanupRedis, err := redispkg.NewRedisClient(redisConfig)
	if err != nil {
		cleanupDB()
		return nil, nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	// 初始化Repositories
	cartRepo := repository.NewCartRepo(db)
	cartItemRepo := repository.NewCartItemRepo(db)

	// 初始化Service
	cartService := service.NewCartService(cartRepo, cartItemRepo, 99, 24)

	cleanup := func() {
		zap.S().Info("cleaning up resources...")
		cleanupRedis()
		cleanupDB()
		cleanupTracer()
	}

	return cartService, cleanup, nil
}

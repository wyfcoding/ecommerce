package main

import (
	"context"
	"fmt"
	"time"

	v1 "ecommerce/api/settlement/v1"
	"ecommerce/internal/settlement/client"
	"ecommerce/internal/settlement/handler"
	"ecommerce/internal/settlement/model"
	"ecommerce/internal/settlement/repository"
	"ecommerce/internal/settlement/service"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	mysqlpkg "ecommerce/pkg/database/mysql"
	redisPkg "ecommerce/pkg/database/redis"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/snowflake"
	"ecommerce/pkg/tracing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm/logger"
)

// Config is the service-specific configuration structure.
type Config struct {
	configpkg.Config
	Data struct {
		configpkg.DataConfig
		Database struct {
			configpkg.DatabaseConfig
			LogLevel      logger.LogLevel `toml:"log_level"`
			SlowThreshold time.Duration     `toml:"slow_threshold"`
		} `toml:"database"`
		Redis redisPkg.Config `toml:"redis"`
		OrderService struct {
			Addr string `toml:"addr"`
		} `toml:"order_service"`
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
	settlementHandler := handler.NewSettlementHandler(srv.(*service.SettlementService))
	// e.g., e.POST("/v1/settlements", settlementHandler.ProcessSettlement)
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)

	// --- Downstream gRPC clients ---
	orderServiceConn, err := grpc.Dial(config.Data.OrderService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to order service: %w", err)
	}

	// --- Data layer ---
	db, cleanupDB, err := mysqlpkg.NewGORMDB(&config.Data.Database)
	if err != nil {
		orderServiceConn.Close()
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	if err := db.AutoMigrate(&model.SettlementRecord{}); err != nil {
		return nil, nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		orderServiceConn.Close()
		return nil, nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	redisClient, cleanupRedis, err := redisPkg.NewRedisClient(&config.Data.Redis)
	if err != nil {
		cleanupDB()
		orderServiceConn.Close()
		return nil, nil, fmt.Errorf("failed to new redis client: %w", err)
	}

	// --- DI (Data -> Biz -> Service) ---
	orderClient := client.NewOrderClient(orderServiceConn)
	settlementRepo := repository.NewSettlementRepo(db, redisClient)
	settlementService := service.NewSettlementService(settlementRepo, orderClient)

	cleanup := func() {
		cleanupRedis()
		cleanupDB()
		orderServiceConn.Close()
	}

	return settlementService, cleanup, nil
}
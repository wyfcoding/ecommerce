package main

import (
	"fmt"
	"time"

	"ecommerce/internal/order/repository"
	"ecommerce/internal/order/service"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	mysqlpkg "ecommerce/pkg/database/mysql"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/tracing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Config 定义了 order-service 的所有配置项
type Config struct {
	configpkg.Config
}

const serviceName = "order-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9093").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	// TODO: 注册gRPC服务
	// v1.RegisterOrderServer(s, srv.(*service.OrderService))
	zap.S().Info("gRPC server registered")
}

func registerGin(e *gin.Engine, srv interface{}) {
	// TODO: 注册HTTP路由
	// handler.RegisterRoutes(e, srv.(*service.OrderService))
	zap.S().Info("HTTP routes registered")
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)

	// 1. 跳过Jaeger初始化
	zap.S().Info("skipping Jaeger tracer initialization...")
	cleanupTracer := func() {}

	// 2. 初始化数据库连接
	zap.S().Info("initializing database connection...")
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
		cleanupTracer()
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 3. 初始化 Repositories
	orderRepo := repository.NewOrderRepo(db)
	orderItemRepo := repository.NewOrderItemRepo(db)
	shippingAddrRepo := repository.NewShippingAddressRepo(db)
	orderLogRepo := repository.NewOrderLogRepo(db)

	cleanupData := func() {}

	// 5. 初始化 Service
	zap.S().Info("initializing order service...")
	orderService := service.NewOrderService(
		orderRepo,
		orderItemRepo,
		shippingAddrRepo,
		orderLogRepo,
		20,  // defaultPageSize
		100, // maxPageSize
		30,  // orderExpirationMinutes
	)

	cleanup := func() {
		zap.S().Info("cleaning up resources...")
		cleanupData()
		cleanupDB()
		cleanupTracer()
	}

	return orderService, cleanup, nil
}

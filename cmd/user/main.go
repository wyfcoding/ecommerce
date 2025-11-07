package main

import (
	"fmt"
	"time"

	"ecommerce/internal/user/handler"
	"ecommerce/internal/user/repository"
	"ecommerce/internal/user/service"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	mysqlpkg "ecommerce/pkg/database/mysql"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/tracing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Config 定义了 user-service 的所有配置项。
type Config struct {
	configpkg.Config
}

const serviceName = "user-service"

// main 是应用程序的入口点。
func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		// 注入 OpenTelemetry (Jaeger) 的拦截器和中间件
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9091"). // 为 metrics server 使用一个不同的端口
		Build().
		Run()
}

// registerGRPC 将 gRPC 服务注册到服务器。
func registerGRPC(s *grpc.Server, srv interface{}) {
	// TODO: 实现 gRPC 服务注册
	// v1.RegisterUserServer(s, srv.(v1.UserServiceServer))
	zap.S().Info("gRPC server registered")
}

// registerGin 将 HTTP 路由注册到 Gin 引擎。
func registerGin(e *gin.Engine, srv interface{}) {
	handler.RegisterRoutes(e, srv.(*service.UserService))
}

// initService 负责初始化服务所需的所有依赖项。
func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)

	// 1. 初始化 Jaeger Tracer (暂时跳过)
	zap.S().Info("skipping Jaeger tracer initialization...")
	cleanupTracer := func() {}
	var err error
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize tracer: %w", err)
	}

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
	userRepo := repository.NewUserRepo(db)
	addressRepo := repository.NewAddressRepo(db)

	cleanupData := func() {}

	// 5. 初始化 Service
	zap.S().Info("initializing user service...")
	userService := service.NewUserService(userRepo, addressRepo, config.JWT.Secret, config.JWT.Issuer, config.JWT.Expire)

	// 定义一个总的清理函数，按初始化的相反顺序执行。
	cleanup := func() {
		zap.S().Info("cleaning up resources...")
		cleanupData()
		cleanupDB()
		cleanupTracer()
	}

	return userService, cleanup, nil
}

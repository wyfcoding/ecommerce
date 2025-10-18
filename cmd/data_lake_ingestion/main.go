package main

import (
	"fmt"
	"time"

	"ecommerce/internal/data_lake_ingestion/handler"

	"ecommerce/internal/data_lake_ingestion/service"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	mysqlpkg "ecommerce/pkg/database/mysql"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/tracing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	v1 "ecommerce/api/data_lake_ingestion/v1"
)

// Config 定义了 data_lake_ingestion-service 的所有配置项。
type Config struct {
	configpkg.Config
	Data struct {
		configpkg.DataConfig
		Database struct {
			configpkg.DatabaseConfig
		} `toml:"database"`
	} `toml:"data"`
	JWT struct {
		Secret string        `toml:"secret"`
		Issuer string        `toml:"issuer"`
		Expire time.Duration `toml:"expire"`
	} `toml:"jwt"`
	Tracing struct {
		JaegerEndpoint string `toml:"jaeger_endpoint"`
	} `toml:"tracing"`
}

const serviceName = "data_lake_ingestion-service"

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
	v1.RegisterDataLakeIngestionServer(s, srv.(v1.DataLakeIngestionServiceServer))
}

// registerGin 将 HTTP 路由注册到 Gin 引擎。
func registerGin(e *gin.Engine, srv interface{}) {
	handler.RegisterRoutes(e, srv.(*service.DataLakeIngestionService))
}

// initService 负责初始化服务所需的所有依赖项。
func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)

	// 1. 初始化 Jaeger Tracer
	zap.S().Info("initializing Jaeger tracer...")
	_, cleanupTracer, err := tracing.InitTracer(&tracing.Config{
		ServiceName:    serviceName,
		JaegerEndpoint: config.Tracing.JaegerEndpoint,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize tracer: %w", err)
	}

	// 2. 初始化数据库连接
	zap.S().Info("initializing database connection...")
	db, cleanupDB, err := mysqlpkg.NewGORMDB(&config.Data.Database)
	if err != nil {
		cleanupTracer()
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 3. 初始化数据层
	// data, cleanupData, err := repository.NewData(db)
	// if err != nil {
	//	cleanupDB()
	//	cleanupTracer()
	//	return nil, nil, fmt.Errorf("failed to initialize data layer: %w", err)
	// }

	// 4. 初始化 Repositories
	// data_lake_ingestionRepo := repository.NewDataLakeIngestionRepo(data)

	// 5. 初始化 Service
	zap.S().Info("initializing data_lake_ingestion service...")
	// data_lake_ingestionService := service.NewDataLakeIngestionService(data_lake_ingestionRepo, config.JWT.Secret, config.JWT.Issuer, config.JWT.Expire)

	// TODO: Replace with actual service initialization
	data_lake_ingestionService := struct{}{
	}

	// 定义一个总的清理函数，按初始化的相反顺序执行。
	cleanup := func() {
		zap.S().Info("cleaning up resources...")
		// cleanupData()
		cleanupDB()
		cleanupTracer()
	}

	return data_lake_ingestionService, cleanup, nil
}
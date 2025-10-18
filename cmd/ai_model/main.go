package main

import (
	"fmt"
	"time"

	v1 "ecommerce/api/ai_model/v1"
	"ecommerce/internal/ai_model/handler"
	"ecommerce/internal/ai_model/service"
	"ecommerce/pkg/app"
	"ecommerce/pkg/config"
	"ecommerce/pkg/database/mysql"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/tracing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Config 定义了 ai_model-service 的所有配置项。
type Config struct {
	config.Config
	Data struct {
		config.DataConfig
		Database struct {
			config.DatabaseConfig
		} `toml:"database"`
	} `toml:"data"`
	Tracing struct {
		JaegerEndpoint string `toml:"jaeger_endpoint"`
	} `toml:"tracing"`
	// 外部服务客户端配置
	Clients struct {
		ProductService struct {
			Addr string `toml:"addr"`
			Port int    `toml:"port"`
		} `toml:"product_service"`
		UserService struct {
			Addr string `toml:"addr"`
			Port int    `toml:"port"`
		} `toml:"user_service"`
		OrderService struct {
			Addr string `toml:"addr"`
			Port int    `toml:"port"`
		} `toml:"order_service"`
		ReviewService struct {
			Addr string `toml:"addr"`
			Port int    `toml:"port"`
		} `toml:"review_service"`
		// TODO: Add other service clients if needed
	} `toml:"clients"`
	AIPlatform struct {
		Endpoint string `toml:"endpoint"` // 外部AI平台或模型服务地址
		ApiKey   string `toml:"api_key"`
	}
}

const serviceName = "ai_model-service"

// main 是应用程序的入口点。
func main() {
	app.NewBuilder(serviceName).WithConfig(&Config{}).WithService(initService).WithGRPC(registerGRPC).WithGin(registerGin).WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).WithMetrics("9091").Build().Run()
}

// registerGRPC 将 gRPC 服务注册到服务器。
func registerGRPC(s *grpc.Server, srv interface{}) {
	v1.RegisterAIModelServiceServer(s, srv.(v1.AIModelServiceServer))
}

// registerGin 将 HTTP 路由注册到 Gin 引擎。
func registerGin(e *gin.Engine, srv interface{}) {
	handler.RegisterRoutes(e, srv.(*service.AIModelService))
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

	// 2. 初始化数据库连接 (如果AI模型服务需要持久化存储，例如模型元数据)
	zap.S().Info("initializing database connection...")
	db, cleanupDB, err := mysqlpkg.NewGORMDB(&config.Data.Database)
	if err != nil {
		cleanupTracer()
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 3. 初始化数据层
	data, cleanupData, err := repository.NewData(db)
	if err != nil {
		cleanupDB()
		cleanupTracer()
		return nil, nil, fmt.Errorf("failed to initialize data layer: %w", err)
	}

	// 4. 初始化 Repositories (如果需要)
	modelMetadataRepo := repository.NewModelMetadataRepo(data)

	// 5. 初始化外部服务客户端
	// Product Service Client
	productServiceAddr := fmt.Sprintf("%s:%d", config.Clients.ProductService.Addr, config.Clients.ProductService.Port)
	productServiceClient, cleanupProductClient, err := client.NewProductServiceClient(productServiceAddr)
	if err != nil {
		cleanupData()
		cleanupDB()
		cleanupTracer()
		return nil, nil, fmt.Errorf("failed to create product service client: %w", err)
	}

	// User Service Client
	userServiceAddr := fmt.Sprintf("%s:%d", config.Clients.UserService.Addr, config.Clients.UserService.Port)
	userServiceClient, cleanupUserClient, err := client.NewUserServiceClient(userServiceAddr)
	if err != nil {
		cleanupProductClient()
		cleanupData()
		cleanupDB()
		cleanupTracer()
		return nil, nil, fmt.Errorf("failed to create user service client: %w", err)
	}

	// Order Service Client
	orderServiceAddr := fmt.Sprintf("%s:%d", config.Clients.OrderService.Addr, config.Clients.OrderService.Port)
	orderServiceClient, cleanupOrderClient, err := client.NewOrderServiceClient(orderServiceAddr)
	if err != nil {
		cleanupUserClient()
		cleanupProductClient()
		cleanupData()
		cleanupDB()
		cleanupTracer()
		return nil, nil, fmt.Errorf("failed to create order service client: %w", err)
	}

	// Review Service Client
	reviewServiceAddr := fmt.Sprintf("%s:%d", config.Clients.ReviewService.Addr, config.Clients.ReviewService.Port)
	reviewServiceClient, cleanupReviewClient, err := client.NewReviewServiceClient(reviewServiceAddr)
	if err != nil {
		cleanupOrderClient()
		cleanupUserClient()
		cleanupProductClient()
		cleanupData()
		cleanupDB()
		cleanupTracer()
		return nil, nil, fmt.Errorf("failed to create review service client: %w", err)
	}

	// 6. 初始化 Service
	zap.S().Info("initializing AI model service...")
	aiModelService := service.NewAIModelService(
		modelMetadataRepo,
		productServiceClient,
		userServiceClient,
		orderServiceClient,
		reviewServiceClient,
		config.AIPlatform.Endpoint,
		config.AIPlatform.ApiKey,
	)

	// 定义一个总的清理函数，按初始化的相反顺序执行。
	cleanup := func() {
		zap.S().Info("cleaning up resources...")
		cleanupReviewClient()
		cleanupOrderClient()
		cleanupUserClient()
		cleanupProductClient()
		cleanupData()
		cleanupDB()
		cleanupTracer()
	}

	return aiModelService, cleanup, nil
}
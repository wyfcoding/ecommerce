package main

import (
	"fmt"
	"time"

	v1 "ecommerce/api/asset/v1"
	"ecommerce/internal/asset/handler"
	"ecommerce/internal/asset/service"
	"ecommerce/pkg/app"
	"ecommerce/pkg/config"
	"ecommerce/pkg/database/mysql"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/tracing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Config 定义了 asset-service 的所有配置项。
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
	ObjectStorage struct {
		Provider string `toml:"provider"` // 例如: "s3", "gcs"
		Bucket   string `toml:"bucket"`
		Region   string `toml:"region"`
		AccessKey string `toml:"access_key"`
		SecretKey string `toml:"secret_key"`
		Endpoint  string `toml:"endpoint"` // MinIO 等兼容 S3 的服务
	} `toml:"object_storage"`
	CDN struct {
		Provider string `toml:"provider"` // 例如: "cloudflare", "aliyun_cdn"
		ApiKey   string `toml:"api_key"`
		ZoneId   string `toml:"zone_id"`
	} `toml:"cdn"`
}

const serviceName = "asset-service"

// main 是应用程序的入口点。
func main() {
	app.NewBuilder(serviceName).WithConfig(&Config{}).WithService(initService).WithGRPC(registerGRPC).WithGin(registerGin).WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).WithMetrics("9091").Build().Run()
}

// registerGRPC 将 gRPC 服务注册到服务器。
func registerGRPC(s *grpc.Server, srv interface{}) {
	v1.RegisterAssetServiceServer(s, srv.(v1.AssetServiceServer))
}

// registerGin 将 HTTP 路由注册到 Gin 引擎。
func registerGin(e *gin.Engine, srv interface{}) {
	handler.RegisterRoutes(e, srv.(*service.AssetService))
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

	// 2. 初始化数据库连接 (用于存储文件元数据)
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

	// 4. 初始化 Repositories
	fileMetadataRepo := repository.NewFileMetadataRepo(data)

	// 5. 初始化对象存储客户端
	objectStorageClient, cleanupObjectStorage, err := client.NewObjectStorageClient(&client.ObjectStorageConfig{
		Provider:  config.ObjectStorage.Provider,
		Bucket:    config.ObjectStorage.Bucket,
		Region:    config.ObjectStorage.Region,
		AccessKey: config.ObjectStorage.AccessKey,
		SecretKey: config.ObjectStorage.SecretKey,
		Endpoint:  config.ObjectStorage.Endpoint,
	})
	if err != nil {
		cleanupData()
		cleanupDB()
		cleanupTracer()
		return nil, nil, fmt.Errorf("failed to create object storage client: %w", err)
	}

	// 6. 初始化 CDN 客户端 (如果配置了)
	var cdnClient client.CDNClient
	var cleanupCDN func()
	if config.CDN.Provider != "" {
		cdnClient, cleanupCDN, err = client.NewCDNClient(&client.CDNConfig{
			Provider: config.CDN.Provider,
			ApiKey:   config.CDN.ApiKey,
			ZoneId:   config.CDN.ZoneId,
		})
		if err != nil {
			cleanupObjectStorage()
			cleanupData()
			cleanupDB()
			cleanupTracer()
			return nil, nil, fmt.Errorf("failed to create CDN client: %w", err)
		}
	}

	// 7. 初始化 Service
	zap.S().Info("initializing asset service...")
	assetService := service.NewAssetService(
		fileMetadataRepo,
		objectStorageClient,
		cdnClient,
	)

	// 定义一个总的清理函数，按初始化的相反顺序执行。
	cleanup := func() {
		zap.S().Info("cleaning up resources...")
		if cleanupCDN != nil {
			cleanupCDN()
		}
		cleanupObjectStorage()
		cleanupData()
		cleanupDB()
		cleanupTracer()
	}

	return assetService, cleanup, nil
}
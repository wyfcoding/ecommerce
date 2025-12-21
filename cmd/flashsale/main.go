package main

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/go-api/flashsale/v1"
	"github.com/wyfcoding/ecommerce/internal/flashsale/application"
	flashCache "github.com/wyfcoding/ecommerce/internal/flashsale/infrastructure/cache"
	"github.com/wyfcoding/ecommerce/internal/flashsale/infrastructure/persistence"
	grpcServer "github.com/wyfcoding/ecommerce/internal/flashsale/interfaces/grpc"
	httpServer "github.com/wyfcoding/ecommerce/internal/flashsale/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
)

// BootstrapName 服务名称常量。
const BootstrapName = "flashsale"

// AppContext 应用上下文，包含配置、服务实例和客户端依赖。
type AppContext struct {
	Config     *configpkg.Config
	AppService *application.FlashsaleService
	Clients    *ServiceClients
}

// ServiceClients 包含所有下游服务的 gRPC 客户端连接。
type ServiceClients struct {
	Product *grpc.ClientConn
}

func main() {
	app.NewBuilder(BootstrapName).
		WithConfig(&configpkg.Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGinMiddleware(middleware.CORS()).
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, svc interface{}) {
	ctx := svc.(*AppContext)
	pb.RegisterFlashSaleServer(s, grpcServer.NewServer(ctx.AppService))
}

func registerGin(e *gin.Engine, svc interface{}) {
	ctx := svc.(*AppContext)
	handler := httpServer.NewHandler(ctx.AppService, slog.Default())
	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	c := cfg.(*configpkg.Config)
	slog.Info("initializing service dependencies...")

	// 1. 数据库
	db, err := databases.NewDB(c.Data.Database, logging.Default())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 2. Redis 缓存
	redisClient, err := cache.NewRedisCache(c.Data.Redis)
	if err != nil {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	// 3. Kafka Producer
	producer := kafka.NewProducer(c.MessageQueue.Kafka, logging.Default())

	// 4. Downstream Clients
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitServiceClients(c.Services, clients)
	if err != nil {
		producer.Close()
		redisClient.Close()
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 5. ID Generator
	idGen, err := idgen.NewSnowflakeGenerator(c.Snowflake)
	if err != nil {
		clientCleanup()
		producer.Close()
		redisClient.Close()
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("failed to initialize id generator: %w", err)
	}

	// 6. Infrastructure & Application
	repo := persistence.NewFlashSaleRepository(db)
	flashSaleCache := flashCache.NewRedisFlashSaleCache(redisClient.GetClient())

	manager := application.NewFlashsaleManager(repo, flashSaleCache, producer, idGen, logging.Default().Logger)
	query := application.NewFlashsaleQuery(repo)
	service := application.NewFlashsaleService(manager, query)

	cleanup := func() {
		slog.Info("cleaning up resources...")
		clientCleanup()
		producer.Close()
		redisClient.Close()
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return &AppContext{
		Config:     c,
		AppService: service,
		Clients:    clients,
	}, cleanup, nil
}

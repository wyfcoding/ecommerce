package main

import (
	"fmt"
	"log/slog"

	"github.com/wyfcoding/pkg/cache"
	"github.com/wyfcoding/pkg/grpcclient"

	pb "github.com/wyfcoding/ecommerce/go-api/order_optimization/v1"
	"github.com/wyfcoding/ecommerce/internal/order_optimization/application"
	"github.com/wyfcoding/ecommerce/internal/order_optimization/infrastructure/persistence"
	optigrpc "github.com/wyfcoding/ecommerce/internal/order_optimization/interfaces/grpc"
	optihttp "github.com/wyfcoding/ecommerce/internal/order_optimization/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

// AppContext 应用上下文，包含配置、服务实例和客户端依赖。
type AppContext struct {
	AppService *application.OrderOptimizationService
	Config     *configpkg.Config
	Clients    *ServiceClients
}

// ServiceClients 包含所有下游服务的 gRPC 客户端连接。
type ServiceClients struct {
	// 如果需要，在此处添加依赖项
}

// BootstrapName 服务名称常量。
const BootstrapName = "order-optimization-service"

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

func registerGRPC(s *grpc.Server, srv interface{}) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	pb.RegisterOrderOptimizationServiceServer(s, optigrpc.NewServer(service))
	slog.Default().Info("gRPC server registered", "service", BootstrapName)
}

func registerGin(e *gin.Engine, srv interface{}) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	handler := optihttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered", "service", BootstrapName)
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	c := cfg.(*configpkg.Config)
	slog.Info("initializing service dependencies...", "service", BootstrapName)

	// 初始化数据库
	db, err := databases.NewDB(c.Data.Database, logging.Default())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	// 初始化 Redis
	redisCache, err := cache.NewRedisCache(c.Data.Redis)
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	// 3. 下游服务客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitServiceClients(c.Services, clients)
	if err != nil {
		redisCache.Close()
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 4. 基础设施与应用层
	repo := persistence.NewOrderOptimizationRepository(db)
	mgr := application.NewOptimizationManager(repo, slog.Default())
	query := application.NewOptimizationQuery(repo)
	service := application.NewOrderOptimizationService(mgr, query)

	cleanup := func() {
		slog.Info("cleaning up resources...")
		clientCleanup()
		redisCache.Close()
		sqlDB.Close()
	}

	return &AppContext{
		Config:     c,
		AppService: service,
		Clients:    clients,
	}, cleanup, nil
}

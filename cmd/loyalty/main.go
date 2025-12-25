package main

import (
	"fmt"
	"log/slog"

	"github.com/wyfcoding/pkg/grpcclient"

	pb "github.com/wyfcoding/ecommerce/goapi/loyalty/v1"
	"github.com/wyfcoding/ecommerce/internal/loyalty/application"
	"github.com/wyfcoding/ecommerce/internal/loyalty/infrastructure/persistence"
	loyaltygrpc "github.com/wyfcoding/ecommerce/internal/loyalty/interfaces/grpc"
	loyaltyhttp "github.com/wyfcoding/ecommerce/internal/loyalty/interfaces/http"
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
	AppService *application.Loyalty
	Config     *configpkg.Config
	Clients    *ServiceClients
}

// ServiceClients 包含所有下游服务的 gRPC 客户端连接。
type ServiceClients struct {
	// 如果需要，在此处添加依赖项
}

// BootstrapName 服务名称常量。
const BootstrapName = "loyalty"

func main() {
	if err := app.NewBuilder(BootstrapName).
		WithConfig(&configpkg.Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGinMiddleware(middleware.CORS()).
		Build().
		Run(); err != nil {
		slog.Error("application run failed", "error", err)
	}
}

func registerGRPC(s *grpc.Server, srv any) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	pb.RegisterLoyaltyServer(s, loyaltygrpc.NewServer(service))
	slog.Default().Info("gRPC server registered (DDD)", "service", BootstrapName)
}

func registerGin(e *gin.Engine, srv any) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	handler := loyaltyhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered (DDD)", "service", BootstrapName)
}

func initService(cfg any, m *metrics.Metrics) (any, func(), error) {
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

	// 基础设施层
	repo := persistence.NewLoyaltyRepository(db)

	// 下游客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, clients)
	if err != nil {
		if sqlDB != nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 应用层
	manager := application.NewLoyaltyManager(repo, slog.Default())
	query := application.NewLoyaltyQuery(repo)
	service := application.NewLoyalty(manager, query)

	cleanup := func() {
		slog.Info("cleaning up resources...")
		clientCleanup()
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

package main

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"github.com/wyfcoding/ecommerce/internal/gateway/application"
	"github.com/wyfcoding/ecommerce/internal/gateway/infrastructure/persistence"
	gatewayhttp "github.com/wyfcoding/ecommerce/internal/gateway/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
)

// AppContext 应用上下文，包含配置、服务实例和客户端依赖。
type AppContext struct {
	AppService *application.GatewayService
	Config     *configpkg.Config
	Clients    *ServiceClients
}

// ServiceClients 包含所有下游服务的 gRPC 客户端连接。
type ServiceClients struct {
	// No dependencies detected
}

// BootstrapName 服务名称常量。
const BootstrapName = "gateway"

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
	slog.Default().Info("gRPC server registered", "service", BootstrapName)
}

func registerGin(e *gin.Engine, srv any) {
	ctx := srv.(*AppContext)
	handler := gatewayhttp.NewHandler(ctx.AppService, slog.Default())

	// 限流和熔断中间件已由 App Builder 根据配置统一添加，无需在此处手动配置

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered", "service", BootstrapName)
}

func initService(cfg any, m *metrics.Metrics) (any, func(), error) {
	c := cfg.(*configpkg.Config)

	// 初始化
	slog.Info("initializing service dependencies...")
	db, err := databases.NewDB(c.Data.Database, logging.Default())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	// 基础设施层
	repo := persistence.NewGatewayRepository(db)

	// 应用层
	service := application.NewGatewayService(repo, slog.Default())

	// 下游客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, clients)
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	cleanup := func() {
		slog.Info("cleaning up resources...")
		clientCleanup()
		slog.Info("cleaning up gateway service resources...")
		sqlDB.Close()
	}

	return &AppContext{
		AppService: service,
		Config:     c,
		Clients:    clients,
	}, cleanup, nil
}

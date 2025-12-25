package main

import (
	"fmt"
	"log/slog"

	"github.com/wyfcoding/pkg/grpcclient"

	pb "github.com/wyfcoding/ecommerce/goapi/risksecurity/v1"
	"github.com/wyfcoding/ecommerce/internal/risksecurity/application"
	"github.com/wyfcoding/ecommerce/internal/risksecurity/infrastructure/persistence"
	riskgrpc "github.com/wyfcoding/ecommerce/internal/risksecurity/interfaces/grpc"
	riskhttp "github.com/wyfcoding/ecommerce/internal/risksecurity/interfaces/http"
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
	AppService *application.RiskService
	Config     *configpkg.Config
	Clients    *ServiceClients
}

// ServiceClients 包含所有下游服务的 gRPC 客户端连接。
type ServiceClients struct {
	// 如果需要，在此处添加依赖项
}

// BootstrapName 服务名称常量。
const BootstrapName = "risksecurity"

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
	pb.RegisterRiskSecurityServer(s, riskgrpc.NewServer(service))
	slog.Default().Info("gRPC server registered", "service", BootstrapName)
}

func registerGin(e *gin.Engine, srv any) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	handler := riskhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered", "service", BootstrapName)
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
	repo := persistence.NewRiskRepository(db)

	// 应用层
	mgr := application.NewRiskManager(repo, logging.Default().Logger)
	query := application.NewRiskQuery(repo)
	service := application.NewRiskService(mgr, query)

	// 下游客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, clients)
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	cleanup := func() {
		slog.Info("cleaning up resources...", "service", BootstrapName)
		clientCleanup()
		sqlDB.Close()
	}

	return &AppContext{
		Config:     c,
		AppService: service,
		Clients:    clients,
	}, cleanup, nil
}

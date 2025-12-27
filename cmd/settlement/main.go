package main

import (
	"fmt"
	"log/slog"

	"github.com/wyfcoding/pkg/grpcclient"

	pb "github.com/wyfcoding/ecommerce/goapi/settlement/v1"
	"github.com/wyfcoding/ecommerce/internal/settlement/application"
	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
	"github.com/wyfcoding/ecommerce/internal/settlement/infrastructure/persistence"
	settlementgrpc "github.com/wyfcoding/ecommerce/internal/settlement/interfaces/grpc"
	settlementhttp "github.com/wyfcoding/ecommerce/internal/settlement/interfaces/http"
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
	AppService *application.SettlementService
	Config     *configpkg.Config
	Clients    *ServiceClients
}

// ServiceClients 包含所有下游服务的 gRPC 客户端连接。
type ServiceClients struct {
	// No dependencies detected
}

// BootstrapName 服务名称常量。
const BootstrapName = "settlement"

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
	pb.RegisterSettlementServer(s, settlementgrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for settlement service (DDD)")
}

func registerGin(e *gin.Engine, srv any) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	handler := settlementhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for settlement service (DDD)")
}

func initService(cfg any, m *metrics.Metrics) (any, func(), error) {
	c := cfg.(*configpkg.Config)
	slog.Info("initializing service dependencies...")

	// 初始化日志
	logger := logging.NewLogger("serviceName", "app")

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
	repo := persistence.NewSettlementRepository(db)
	ledgerRepo := persistence.NewLedgerRepository(db)

	// 领域/服务层
	ledgerService := domain.NewLedgerService(ledgerRepo)

	// 3. 下游服务客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, clients)
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 4. 基础设施与应用层 & Domain
	// 注意：ledgerService 已在上方初始化
	service := application.NewSettlementService(repo, ledgerService, logger.Logger)

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

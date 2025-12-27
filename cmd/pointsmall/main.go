package main

import (
	"fmt"
	"log/slog"

	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/idgen"

	pb "github.com/wyfcoding/ecommerce/goapi/pointsmall/v1"
	"github.com/wyfcoding/ecommerce/internal/pointsmall/application"
	"github.com/wyfcoding/ecommerce/internal/pointsmall/infrastructure/persistence"
	pointsgrpc "github.com/wyfcoding/ecommerce/internal/pointsmall/interfaces/grpc"
	pointshttp "github.com/wyfcoding/ecommerce/internal/pointsmall/interfaces/http"
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
	AppService *application.PointsmallService
	Config     *configpkg.Config
	Clients    *ServiceClients
}

// ServiceClients 包含所有下游服务的 gRPC 客户端连接。
type ServiceClients struct {
	// No dependencies detected
}

// BootstrapName 服务名称常量。
const BootstrapName = "pointsmall"

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
	pb.RegisterPointsmallServer(s, pointsgrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for pointsmall service (DDD)")
}

func registerGin(e *gin.Engine, srv any) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	handler := pointshttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for pointsmall service (DDD)")
}

func initService(cfg any, m *metrics.Metrics) (any, func(), error) {
	c := cfg.(*configpkg.Config)
	slog.Info("initializing service dependencies...")

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
	// 原始 repo 声明: repo := persistence.NewPointsRepository(db)

	// 3. 下游服务客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, clients)
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 4. 基础设施与应用层
	// 初始化 ID 生成器
	idGen, err := idgen.NewSnowflakeGenerator(configpkg.SnowflakeConfig{
		MachineID: 1,
	})
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to initialize id generator: %w", err)
	}

	// 4. 基础设施与应用层
	repo := persistence.NewPointsRepository(db)
	mgr := application.NewPointsManager(repo, idGen, logging.Default().Logger)
	query := application.NewPointsQuery(repo)
	service := application.NewPointsmallService(mgr, query)

	cleanup := func() {
		slog.Info("cleaning up resources...")
		clientCleanup()
		sqlDB.Close()
	}

	return &AppContext{
		Config:     c,
		AppService: service,
		Clients:    clients,
	}, cleanup, nil
}

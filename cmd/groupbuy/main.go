package main

import (
	"fmt"
	"log/slog"

	"github.com/wyfcoding/pkg/grpcclient"

	pb "github.com/wyfcoding/ecommerce/goapi/groupbuy/v1"
	"github.com/wyfcoding/ecommerce/internal/groupbuy/application"
	"github.com/wyfcoding/ecommerce/internal/groupbuy/infrastructure/persistence"
	groupbuygrpc "github.com/wyfcoding/ecommerce/internal/groupbuy/interfaces/grpc"
	groupbuyhttp "github.com/wyfcoding/ecommerce/internal/groupbuy/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

// AppContext 应用上下文，包含配置、服务实例和客户端依赖。
type AppContext struct {
	AppService *application.GroupbuyService
	Config     *configpkg.Config
	Clients    *ServiceClients
}

// ServiceClients 包含所有下游服务的 gRPC 客户端连接。
type ServiceClients struct {
	// 如果需要，在此处添加依赖项
}

// BootstrapName 服务名称常量。
const BootstrapName = "groupbuy"

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
	pb.RegisterGroupbuyServer(s, groupbuygrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for groupbuy service (DDD)")
}

func registerGin(e *gin.Engine, srv any) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	handler := groupbuyhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for groupbuy service (DDD)")
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
	repo := persistence.NewGroupbuyRepository(db)

	// 下游客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, clients)
	if err != nil {
		if sqlDB != nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// ID 生成器
	idGenerator, err := idgen.NewSnowflakeGenerator(c.Snowflake)
	if err != nil {
		clientCleanup()
		if sqlDB != nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("failed to create id generator: %w", err)
	}

	// 应用层
	manager := application.NewGroupbuyManager(repo, idGenerator, slog.Default())
	query := application.NewGroupbuyQuery(repo)
	service := application.NewGroupbuyService(manager, query)

	cleanup := func() {
		slog.Default().Info("cleaning up groupbuy service resources (DDD)...")
		clientCleanup()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return &AppContext{AppService: service, Config: c, Clients: clients}, cleanup, nil
}

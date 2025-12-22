package main

import (
	"fmt"
	"log/slog"

	"github.com/wyfcoding/pkg/grpcclient"

	pb "github.com/wyfcoding/ecommerce/go-api/recommendation/v1"
	"github.com/wyfcoding/ecommerce/internal/recommendation/application"
	"github.com/wyfcoding/ecommerce/internal/recommendation/infrastructure/persistence"
	recommgrpc "github.com/wyfcoding/ecommerce/internal/recommendation/interfaces/grpc"
	recommhttp "github.com/wyfcoding/ecommerce/internal/recommendation/interfaces/http"
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
	AppService *application.RecommendationService
	Config     *configpkg.Config
	Clients    *ServiceClients
}

// ServiceClients 包含所有下游服务的 gRPC 客户端连接。
type ServiceClients struct {
	User    *grpc.ClientConn
	Product *grpc.ClientConn
	Order   *grpc.ClientConn
}

// BootstrapName 服务名称常量。
const BootstrapName = "recommendation-service"

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

func registerGRPC(s *grpc.Server, srv any) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	pb.RegisterRecommendationServiceServer(s, recommgrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for recommendation service (DDD)")
}

func registerGin(e *gin.Engine, srv any) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	handler := recommhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for recommendation service (DDD)")
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
	repo := persistence.NewRecommendationRepository(db)

	// 下游客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitServiceClients(c.Services, clients)
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 应用层
	query := application.NewRecommendationQuery(repo)
	manager := application.NewRecommendationManager(repo, logging.Default().Logger)
	service := application.NewRecommendationService(manager, query)

	cleanup := func() {
		slog.Info("cleaning up resources...")
		clientCleanup()
		sqlDB.Close()
	}

	return &AppContext{AppService: service, Config: c, Clients: clients}, cleanup, nil
}

package main

import (
	"fmt" // 保留此行，因为指令是移除不存在的 'reflect'。提供的 'Code Edit' 片段有误导性。
	"log/slog"

	"github.com/wyfcoding/pkg/grpcclient"

	pb "github.com/wyfcoding/ecommerce/go-api/logistics_routing/v1"
	"github.com/wyfcoding/ecommerce/internal/logistics_routing/application"
	"github.com/wyfcoding/ecommerce/internal/logistics_routing/infrastructure/persistence"
	routinggrpc "github.com/wyfcoding/ecommerce/internal/logistics_routing/interfaces/grpc"
	routinghttp "github.com/wyfcoding/ecommerce/internal/logistics_routing/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type AppContext struct {
	AppService *application.LogisticsRoutingService
	Config     *configpkg.Config
	Clients    *ServiceClients
}

type ServiceClients struct {
	// 如果需要，在此处添加依赖项
}

const BootstrapName = "logistics-routing-service"

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
	pb.RegisterLogisticsRoutingServiceServer(s, routinggrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for logistics_routing service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	handler := routinghttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for logistics_routing service (DDD)")
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	c := cfg.(*configpkg.Config)
	slog.Info("initializing service dependencies...", "service", BootstrapName)

	logging.NewLogger(BootstrapName, "app")

	db, err := databases.NewDB(c.Data.Database, logging.Default())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	repo := persistence.NewLogisticsRoutingRepository(db)

	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitServiceClients(c.Services, clients)
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	mgr := application.NewLogisticsRoutingManager(repo, slog.Default())
	query := application.NewLogisticsRoutingQuery(repo)
	service := application.NewLogisticsRoutingService(mgr, query)

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

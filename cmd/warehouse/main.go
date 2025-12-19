package main

import (
	"fmt"
	"log/slog"

	"github.com/wyfcoding/pkg/grpcclient"

	pb "github.com/wyfcoding/ecommerce/go-api/warehouse/v1"
	"github.com/wyfcoding/ecommerce/internal/warehouse/application"
	"github.com/wyfcoding/ecommerce/internal/warehouse/infrastructure/persistence"
	warehousegrpc "github.com/wyfcoding/ecommerce/internal/warehouse/interfaces/grpc"
	warehousehttp "github.com/wyfcoding/ecommerce/internal/warehouse/interfaces/http"
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
	AppService *application.WarehouseService
	Config     *configpkg.Config
	Clients    *ServiceClients
}

type ServiceClients struct {
	// Add dependencies here if needed
}

const BootstrapName = "warehouse-service"

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
	pb.RegisterWarehouseServiceServer(s, warehousegrpc.NewServer(service))
	slog.Default().Info("gRPC server registered (DDD)", "service", BootstrapName)
}

func registerGin(e *gin.Engine, srv interface{}) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	handler := warehousehttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered (DDD)", "service", BootstrapName)
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	c := cfg.(*configpkg.Config)

	// Initialize Logger
	logger := logging.NewLogger("serviceName", "app")

	// Initialize
	slog.Info("initializing service dependencies...")
	db, err := databases.NewDB(c.Data.Database, logging.Default())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	// Infrastructure Layer
	// 3. Downstream Clients
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitServiceClients(c.Services, clients)
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 4. Infrastructure & Application
	repo := persistence.NewWarehouseRepository(db)
	query := application.NewWarehouseQuery(repo)
	manager := application.NewWarehouseManager(repo, logger.Logger)
	service := application.NewWarehouseService(manager, query)

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

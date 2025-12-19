package main

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/go-api/inventory/v1"
	"github.com/wyfcoding/ecommerce/internal/inventory/application"
	"github.com/wyfcoding/ecommerce/internal/inventory/infrastructure/persistence"
	inventorygrpc "github.com/wyfcoding/ecommerce/internal/inventory/interfaces/grpc"
	inventoryhttp "github.com/wyfcoding/ecommerce/internal/inventory/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
)

const BootstrapName = "inventory"

type AppContext struct {
	Config     *configpkg.Config
	AppService *application.InventoryService
	Clients    *ServiceClients
}

type ServiceClients struct {
	Warehouse         *grpc.ClientConn
	Logistics         *grpc.ClientConn
	Product           *grpc.ClientConn
	InventoryForecast *grpc.ClientConn
}

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

func registerGRPC(s *grpc.Server, svc interface{}) {
	ctx := svc.(*AppContext)
	pb.RegisterInventoryServiceServer(s, inventorygrpc.NewServer(ctx.AppService))
}

func registerGin(e *gin.Engine, svc interface{}) {
	ctx := svc.(*AppContext)
	handler := inventoryhttp.NewHandler(ctx.AppService, slog.Default())
	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	c := cfg.(*configpkg.Config)
	slog.Info("initializing service dependencies...")

	// 1. Database
	db, err := databases.NewDB(c.Data.Database, logging.Default())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 2. Redis
	redisCache, err := cache.NewRedisCache(c.Data.Redis)
	if err != nil {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	// 3. Downstream Clients
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitServiceClients(c.Services, clients)
	if err != nil {
		redisCache.Close()
		sqlDB, _ := db.DB()
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 4. Infrastructure & Application
	repo := persistence.NewInventoryRepository(db)
	warehouseRepo := persistence.NewWarehouseRepository(db)

	// Create sub-services
	inventoryQuery := application.NewInventoryQuery(repo, warehouseRepo, logging.Default().Logger)
	inventoryManager := application.NewInventoryManager(repo, warehouseRepo, logging.Default().Logger)

	// Create facade
	service := application.NewInventoryService(inventoryManager, inventoryQuery)

	cleanup := func() {
		slog.Info("cleaning up resources...")
		clientCleanup()
		redisCache.Close()
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	return &AppContext{
		Config:     c,
		AppService: service,
		Clients:    clients,
	}, cleanup, nil
}

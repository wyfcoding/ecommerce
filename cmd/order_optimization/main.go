package main

import (
	"fmt"
	"log/slog"

	"github.com/wyfcoding/pkg/cache"
	"github.com/wyfcoding/pkg/grpcclient"

	pb "github.com/wyfcoding/ecommerce/go-api/order_optimization/v1"
	"github.com/wyfcoding/ecommerce/internal/order_optimization/application"
	"github.com/wyfcoding/ecommerce/internal/order_optimization/infrastructure/persistence"
	optigrpc "github.com/wyfcoding/ecommerce/internal/order_optimization/interfaces/grpc"
	optihttp "github.com/wyfcoding/ecommerce/internal/order_optimization/interfaces/http"
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
	AppService *application.OrderOptimizationService
	Config     *configpkg.Config
	Clients    *ServiceClients
}

type ServiceClients struct {
	// Add dependencies here if needed
}

const BootstrapName = "order-optimization-service"

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
	pb.RegisterOrderOptimizationServiceServer(s, optigrpc.NewServer(service))
	slog.Default().Info("gRPC server registered", "service", BootstrapName)
}

func registerGin(e *gin.Engine, srv interface{}) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	handler := optihttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered", "service", BootstrapName)
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	c := cfg.(*configpkg.Config)
	slog.Info("initializing service dependencies...", "service", BootstrapName)

	// Initialize Database
	db, err := databases.NewDB(c.Data.Database, logging.Default())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	// Initialize Redis
	redisCache, err := cache.NewRedisCache(c.Data.Redis)
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	// 3. Downstream Clients
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitServiceClients(c.Services, clients)
	if err != nil {
		redisCache.Close()
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 4. Infrastructure & Application
	repo := persistence.NewOrderOptimizationRepository(db)
	service := application.NewOrderOptimizationService(repo, logging.Default().Logger)

	cleanup := func() {
		slog.Info("cleaning up resources...")
		clientCleanup()
		redisCache.Close()
		sqlDB.Close()
	}

	return &AppContext{
		Config:     c,
		AppService: service,
		Clients:    clients,
	}, cleanup, nil
}

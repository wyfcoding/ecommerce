package main

import (
	"fmt"
	"log/slog"

	"github.com/wyfcoding/pkg/grpcclient"

	pb "github.com/wyfcoding/ecommerce/go-api/subscription/v1"
	"github.com/wyfcoding/ecommerce/internal/subscription/application"
	"github.com/wyfcoding/ecommerce/internal/subscription/infrastructure/persistence"
	subscriptiongrpc "github.com/wyfcoding/ecommerce/internal/subscription/interfaces/grpc"
	subscriptionhttp "github.com/wyfcoding/ecommerce/internal/subscription/interfaces/http"
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
	AppService *application.SubscriptionService
	Config     *configpkg.Config
	Clients    *ServiceClients
}

type ServiceClients struct {
	// Add dependencies here if needed
}

const BootstrapName = "subscription-service"

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
	pb.RegisterSubscriptionServiceServer(s, subscriptiongrpc.NewServer(service))
	slog.Default().Info("gRPC server registered", "service", BootstrapName) // (DDD)
}

func registerGin(e *gin.Engine, srv interface{}) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	handler := subscriptionhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered", "service", BootstrapName) // (DDD)
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

	// Infrastructure Layer
	repo := persistence.NewSubscriptionRepository(db)

	// Application Layer
	mgr := application.NewSubscriptionManager(repo, logging.Default().Logger)
	query := application.NewSubscriptionQuery(repo)
	service := application.NewSubscriptionService(mgr, query)

	// Downstream Clients
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitServiceClients(c.Services, clients)
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

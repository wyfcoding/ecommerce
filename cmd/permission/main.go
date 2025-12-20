package main

import (
	"fmt"
	"log/slog"

	"github.com/wyfcoding/pkg/grpcclient"

	"github.com/gin-gonic/gin"
	pb "github.com/wyfcoding/ecommerce/go-api/permission/v1"
	"github.com/wyfcoding/ecommerce/internal/permission/application"
	"github.com/wyfcoding/ecommerce/internal/permission/infrastructure/persistence"
	permissiongrpc "github.com/wyfcoding/ecommerce/internal/permission/interfaces/grpc"
	permissionhttp "github.com/wyfcoding/ecommerce/internal/permission/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"

	"google.golang.org/grpc"
)

const BootstrapName = "permission-service"

type AppContext struct {
	AppService *application.PermissionService
	Config     *configpkg.Config
	Clients    *ServiceClients
}

type ServiceClients struct {
	// Add dependencies here if needed
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

func registerGRPC(s *grpc.Server, srv interface{}) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	pb.RegisterPermissionServiceServer(s, permissiongrpc.NewServer(service))
	slog.Default().Info("gRPC server registered (DDD)", "service", BootstrapName)
}

func registerGin(e *gin.Engine, srv interface{}) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	handler := permissionhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered (DDD)", "service", BootstrapName)
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	c := cfg.(*configpkg.Config)
	slog.Info("initializing service dependencies...", "service", BootstrapName)

	// Initialize Logger
	logger := logging.NewLogger(BootstrapName, "app")

	// Initialize Database
	db, err := databases.NewDB(c.Data.Database, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// Infrastructure Layer
	repo := persistence.NewPermissionRepository(db)

	// 3. Downstream Clients
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitServiceClients(c.Services, clients)
	if err != nil {
		// sqlDB.Close() // sqlDB is accessed via db.DB() if needed, but here we just return
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 4. Infrastructure & Application
	mgr := application.NewPermissionManager(repo, slog.Default())
	query := application.NewPermissionQuery(repo)
	service := application.NewPermissionService(mgr, query)

	cleanup := func() {
		slog.Info("cleaning up resources...")
		clientCleanup()
		// sqlDB.Close()
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	return &AppContext{
		Config:     c,
		AppService: service,
		Clients:    clients,
	}, cleanup, nil
}

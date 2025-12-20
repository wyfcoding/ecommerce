package main

import (
	"fmt"
	"log/slog"

	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/idgen"

	pb "github.com/wyfcoding/ecommerce/go-api/pointsmall/v1"
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

type AppContext struct {
	AppService *application.PointsmallService
	Config     *configpkg.Config
	Clients    *ServiceClients
}

type ServiceClients struct {
	// Add dependencies here if needed
}

const BootstrapName = "pointsmall-service"

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
	pb.RegisterPointsmallServiceServer(s, pointsgrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for pointsmall service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	handler := pointshttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for pointsmall service (DDD)")
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	c := cfg.(*configpkg.Config)
	slog.Info("initializing service dependencies...")

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
	// Original repo declaration: repo := persistence.NewPointsRepository(db)

	// 3. Downstream Clients
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitServiceClients(c.Services, clients)
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 4. Infrastructure & Application
	// Initialize ID Generator
	idGen, err := idgen.NewSnowflakeGenerator(configpkg.SnowflakeConfig{
		MachineID: 1,
	})
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to initialize id generator: %w", err)
	}

	// 4. Infrastructure & Application
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

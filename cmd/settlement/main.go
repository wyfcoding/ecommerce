package main

import (
	"fmt"
	"log/slog"

	"github.com/wyfcoding/pkg/grpcclient"

	pb "github.com/wyfcoding/ecommerce/go-api/settlement/v1"
	"github.com/wyfcoding/ecommerce/internal/settlement/application"
	"github.com/wyfcoding/ecommerce/internal/settlement/domain/ledger"
	"github.com/wyfcoding/ecommerce/internal/settlement/infrastructure/persistence"
	settlementgrpc "github.com/wyfcoding/ecommerce/internal/settlement/interfaces/grpc"
	settlementhttp "github.com/wyfcoding/ecommerce/internal/settlement/interfaces/http"
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
	AppService *application.SettlementService
	Config     *configpkg.Config
	Clients    *ServiceClients
}

type ServiceClients struct {
	// Add dependencies here if needed
}

const BootstrapName = "settlement-service"

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
	pb.RegisterSettlementServiceServer(s, settlementgrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for settlement service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	handler := settlementhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for settlement service (DDD)")
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	c := cfg.(*configpkg.Config)
	slog.Info("initializing service dependencies...")

	// Initialize Logger
	logger := logging.NewLogger("serviceName", "app")

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
	repo := persistence.NewSettlementRepository(db)
	ledgerRepo := persistence.NewLedgerRepository(db)

	// Domain/Service Layer
	ledgerService := ledger.NewLedgerService(ledgerRepo)

	// 3. Downstream Clients
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitServiceClients(c.Services, clients)
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 4. Infrastructure & Application & Domain
	// Note: ledgerService is already initialized above
	service := application.NewSettlementService(repo, ledgerService, logger.Logger)

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

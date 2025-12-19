package main

import (
	"fmt"
	"log/slog"

	"github.com/wyfcoding/pkg/grpcclient"

	pb "github.com/wyfcoding/ecommerce/go-api/message/v1"
	"github.com/wyfcoding/ecommerce/internal/message/application"
	"github.com/wyfcoding/ecommerce/internal/message/infrastructure/persistence"
	messagegrpc "github.com/wyfcoding/ecommerce/internal/message/interfaces/grpc"
	messagehttp "github.com/wyfcoding/ecommerce/internal/message/interfaces/http"
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
	AppService *application.MessageService
	Config     *configpkg.Config
	Clients    *ServiceClients
}

type ServiceClients struct {
	// Add dependencies here if needed
}

const BootstrapName = "message-service"

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
	pb.RegisterMessageServiceServer(s, messagegrpc.NewServer(service))
	slog.Default().Info("gRPC server registered (DDD)", "service", BootstrapName)
}

func registerGin(e *gin.Engine, srv interface{}) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	handler := messagehttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered (DDD)", "service", BootstrapName)
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	c := cfg.(*configpkg.Config)

	// Initialize Logger
	logging.NewLogger(BootstrapName, "app") // Set default logger for the app

	slog.Info("initializing service dependencies...", "service", BootstrapName)
	db, err := databases.NewDB(c.Data.Database, logging.Default())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	// 3. Downstream Clients
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitServiceClients(c.Services, clients)
	if err != nil {
		// Assuming redisCache is not present in this context, removing its cleanup
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 4. Infrastructure & Application
	repo := persistence.NewMessageRepository(db)
	service := application.NewMessageService(repo, logging.Default().Logger)

	cleanup := func() {
		slog.Info("cleaning up resources...")
		clientCleanup()
		// Assuming redisCache is not present in this context, removing its cleanup
		sqlDB.Close()
	}

	return &AppContext{
		Config:     c,
		AppService: service,
		Clients:    clients,
	}, cleanup, nil
}

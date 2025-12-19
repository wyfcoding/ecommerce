package main

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/go-api/notification/v1"
	"github.com/wyfcoding/ecommerce/internal/notification/application"
	"github.com/wyfcoding/ecommerce/internal/notification/infrastructure/persistence"
	notifgrpc "github.com/wyfcoding/ecommerce/internal/notification/interfaces/grpc"
	notifhttp "github.com/wyfcoding/ecommerce/internal/notification/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
)

const BootstrapName = "notification"

type AppContext struct {
	Config     *configpkg.Config
	AppService *application.NotificationService
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

func registerGRPC(s *grpc.Server, svc interface{}) {
	ctx := svc.(*AppContext)
	pb.RegisterNotificationServiceServer(s, notifgrpc.NewServer(ctx.AppService))
}

func registerGin(e *gin.Engine, svc interface{}) {
	ctx := svc.(*AppContext)
	handler := notifhttp.NewHandler(ctx.AppService, slog.Default())
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
	repo := persistence.NewNotificationRepository(db)

	// Create sub-services
	notifQuery := application.NewNotificationQuery(repo, logging.Default().Logger)
	notifManager := application.NewNotificationManager(repo, logging.Default().Logger)

	// Create facade
	service := application.NewNotificationService(notifManager, notifQuery)

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

package main

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/go-api/user/v1"
	"github.com/wyfcoding/ecommerce/internal/user/application"
	mysqlRepo "github.com/wyfcoding/ecommerce/internal/user/infrastructure/persistence/mysql"
	usergrpc "github.com/wyfcoding/ecommerce/internal/user/interfaces/grpc"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
)

const BootstrapName = "user"

type AppContext struct {
	Config     *configpkg.Config
	AppService *application.UserService
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
	pb.RegisterUserServer(s, usergrpc.NewServer(ctx.AppService))
}

func registerGin(e *gin.Engine, svc interface{}) {
	// User service currently has no HTTP routes in the original code,
	// but we keep the callback for future use or health checks.
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
	repo := mysqlRepo.NewUserRepository(db)
	addressRepo := mysqlRepo.NewAddressRepository(db)

	// Services
	logger := logging.Default().Logger
	authService := application.NewAuthService(repo, c.JWT.Secret, c.JWT.Issuer, c.JWT.ExpireDuration, logger)
	profileService := application.NewProfileService(repo, logger)
	addressService := application.NewAddressService(repo, addressRepo, logger)

	// Facade
	svc := application.NewUserService(
		authService,
		profileService,
		addressService,
		logger,
	)

	cleanup := func() {
		slog.Info("cleaning up resources...")
		clientCleanup()
		redisCache.Close()
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	return &AppContext{
		Config:     c,
		AppService: svc,
		Clients:    clients,
	}, cleanup, nil
}

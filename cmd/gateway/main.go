package main

import (
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/gateway/application"
	"github.com/wyfcoding/ecommerce/internal/gateway/infrastructure/persistence"
	gatewayhttp "github.com/wyfcoding/ecommerce/internal/gateway/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

// AppContext 应用上下文，包含配置、服务实例和客户端依赖。
type AppContext struct {
	AppService *application.GatewayService
	Config     *configpkg.Config
	Clients    *ServiceClients
}

// ServiceClients 包含所有下游服务的 gRPC 客户端连接。
type ServiceClients struct {
	User              *grpc.ClientConn
	Product           *grpc.ClientConn
	Order             *grpc.ClientConn
	Cart              *grpc.ClientConn
	Payment           *grpc.ClientConn
	Inventory         *grpc.ClientConn
	Marketing         *grpc.ClientConn
	Notification      *grpc.ClientConn
	Search            *grpc.ClientConn
	Recommendation    *grpc.ClientConn
	Review            *grpc.ClientConn
	Wishlist          *grpc.ClientConn
	Logistics         *grpc.ClientConn
	Aftersales        *grpc.ClientConn
	CustomerService   *grpc.ClientConn
	ContentModeration *grpc.ClientConn
	Message           *grpc.ClientConn
	File              *grpc.ClientConn
}

// BootstrapName 服务名称常量。
const BootstrapName = "gateway"

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
	slog.Default().Info("gRPC server registered", "service", BootstrapName)
}

func registerGin(e *gin.Engine, srv interface{}) {
	ctx := srv.(*AppContext)
	cfg := ctx.Config
	handler := gatewayhttp.NewHandler(ctx.AppService, slog.Default())

	// 根据配置应用中间件
	if cfg.RateLimit.Enabled {
		slog.Info("Enabling Rate Limit Middleware", "rate", cfg.RateLimit.Rate, "burst", cfg.RateLimit.Burst)
		e.Use(middleware.NewLocalRateLimitMiddleware(int(cfg.RateLimit.Rate), cfg.RateLimit.Burst))
	}
	if cfg.CircuitBreaker.Enabled {
		slog.Info("Enabling Circuit Breaker Middleware")
		e.Use(middleware.CircuitBreaker())
	}

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered", "service", BootstrapName)
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	c := cfg.(*configpkg.Config)

	// 初始化日志

	// 初始化
	slog.Info("initializing service dependencies...")
	db, err := databases.NewDB(c.Data.Database, logging.Default())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	// 基础设施层
	repo := persistence.NewGatewayRepository(db)

	// 应用层
	service := application.NewGatewayService(repo, slog.Default())

	// 下游客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitServiceClients(c.Services, clients)
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	cleanup := func() {
		slog.Info("cleaning up resources...")
		clientCleanup()
		slog.Info("cleaning up gateway service resources...")
		sqlDB.Close()
	}

	return &AppContext{
			AppService: service,
			Config:     c,
			Clients:    clients,
		}, func() {
			cleanup()
		}, nil
}

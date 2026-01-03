package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"github.com/wyfcoding/ecommerce/internal/gateway/application"
	"github.com/wyfcoding/ecommerce/internal/gateway/infrastructure/k8s"
	"github.com/wyfcoding/ecommerce/internal/gateway/infrastructure/persistence"
	gatewayhttp "github.com/wyfcoding/ecommerce/internal/gateway/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/limiter"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// BootstrapName 服务标识。
const BootstrapName = "gateway"

// Config 扩展基础配置。
type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

// AppContext 应用资源上下文。
type AppContext struct {
	Config     *Config
	AppService *application.GatewayService
	Clients    *ServiceClients
	Metrics    *metrics.Metrics
	Limiter    limiter.Limiter
}

// ServiceClients 下游微服务。
type ServiceClients struct{}

func main() {
	if err := app.NewBuilder(BootstrapName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGinMiddleware(
			middleware.CORS(),
		).
		Build().
		Run(); err != nil {
		slog.Error("service bootstrap failed", "error", err)
	}
}

func registerGRPC(s *grpc.Server, srv any) {
	slog.Info("gRPC handler registered", "service", BootstrapName)
}

func registerGin(e *gin.Engine, srv any) {
	ctx := srv.(*AppContext)

	if ctx.Config.Server.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	e.Use(middleware.RateLimitWithLimiter(ctx.Limiter))

	// 3. 业务路由注册
	handler := gatewayhttp.NewHandler(ctx.AppService, slog.Default())
	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)
}

func initService(cfg any, m *metrics.Metrics) (any, func(), error) {
	c := cfg.(*Config)
	logger := logging.Default()

	db, err := databases.NewDB(c.Data.Database, c.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("database init failed: %w", err)
	}
	slog.Info("Database initialized", "driver", c.Data.Database.Driver)

	redisCache, err := cache.NewRedisCache(c.Data.Redis, c.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("redis init failed: %w", err)
	}
	slog.Info("Redis cache initialized", "addr", c.Data.Redis.Addr)

	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), c.RateLimit.Rate, time.Second)
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, m, c.CircuitBreaker, clients)
	if err != nil {
		return nil, nil, fmt.Errorf("grpc clients init failed: %w", err)
	}
	slog.Info("GRPC clients initialized", "count", len(c.Services))

	// K8s 控制器
	var k8sConfig *rest.Config
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig != "" {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		k8sConfig, err = rest.InClusterConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())
	repo := persistence.NewGatewayRepository(db.RawDB())
	service := application.NewGatewayService(repo, logger.Logger)

	if err == nil {
		dynamicClient, err := dynamic.NewForConfig(k8sConfig)
		if err != nil {
			slog.Error("failed to create dynamic k8s client", "error", err)
		} else {
			controller := k8s.NewRouteController(dynamicClient, service, logger.Logger)
			go func() {
				slog.Info("Starting k8s route controller...")
				if err := controller.Start(ctx); err != nil {
					slog.Error("failed to start k8s route controller", "error", err)
				}
			}()
		}
	} else {
		slog.Warn("skipping k8s controller initialization: k8s config not found or invalid", "error", err)
	}

	cleanup := func() {
		slog.Info("Service shutting down, cleaning up resources...")
		cancel()
		clientCleanup()
		redisCache.Close()
		slog.Info("Cleanup completed")
	}

	return &AppContext{
		Config:     c,
		AppService: service,
		Clients:    clients,
		Metrics:    m,
		Limiter:    rateLimiter,
	}, cleanup, nil
}
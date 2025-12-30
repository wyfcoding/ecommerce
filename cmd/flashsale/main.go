package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/goapi/flashsale/v1"
	"github.com/wyfcoding/ecommerce/internal/flashsale/application"
	flashCache "github.com/wyfcoding/ecommerce/internal/flashsale/infrastructure/cache"
	"github.com/wyfcoding/ecommerce/internal/flashsale/infrastructure/persistence"
	grpcServer "github.com/wyfcoding/ecommerce/internal/flashsale/interfaces/grpc"
	httpServer "github.com/wyfcoding/ecommerce/internal/flashsale/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/limiter"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
)

// BootstrapName 服务标识。
const BootstrapName = "flashsale"

// Config 扩展配置结构。
type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

// AppContext 应用资源上下文。
type AppContext struct {
	Config     *Config
	AppService *application.FlashsaleService
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
			middleware.MetricsMiddleware(),
			middleware.CORS(),
		).
		Build().
		Run(); err != nil {
		slog.Error("service bootstrap failed", "error", err)
	}
}

func registerGRPC(s *grpc.Server, svc any) {
	ctx := svc.(*AppContext)
	pb.RegisterFlashSaleServer(s, grpcServer.NewServer(ctx.AppService))
}

func registerGin(e *gin.Engine, svc any) {
	ctx := svc.(*AppContext)

	if ctx.Config.Server.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 1. 系统路由组 (不限流)
	sys := e.Group("/sys")
	{
		sys.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "UP",
				"service":   BootstrapName,
				"timestamp": time.Now().Unix(),
			})
		})
		sys.GET("/ready", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "READY"})
		})
	}

	if ctx.Config.Metrics.Enabled {
		e.GET(ctx.Config.Metrics.Path, gin.WrapH(ctx.Metrics.Handler()))
	}

	// 2. 治理：限流保护
	e.Use(middleware.RateLimitWithLimiter(ctx.Limiter))

	// 3. 业务路由
	handler := httpServer.NewHandler(ctx.AppService, slog.Default())
	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Info("HTTP service configured successfully", "service", BootstrapName)
}

func initService(cfg any, m *metrics.Metrics) (any, func(), error) {
	c := cfg.(*Config)
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default()

	// 1. 基础设施
	db, err := databases.NewDB(c.Data.Database, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("database init failed: %w", err)
	}

	redisCache, err := cache.NewRedisCache(c.Data.Redis)
	if err != nil {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("redis init failed: %w", err)
	}

	// 2. 治理：分布式限流器
	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), c.RateLimit.Rate, time.Second)

	// 3. 消息队列与 ID 生成器
	producer := kafka.NewProducer(c.MessageQueue.Kafka, logger)
	idGen, err := idgen.NewSnowflakeGenerator(c.Snowflake)
	if err != nil {
		producer.Close()
		redisCache.Close()
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("idgen init failed: %w", err)
	}

	// 4. 下游微服务拨号
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, clients)
	if err != nil {
		producer.Close()
		redisCache.Close()
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("grpc clients init failed: %w", err)
	}

	// 5. DDD 分层装配
	bootLog.Info("assembling flashsale domain services...")
	repo := persistence.NewFlashSaleRepository(db)
	flashSaleCache := flashCache.NewRedisFlashSaleCache(redisCache.GetClient())
	manager := application.NewFlashsaleManager(repo, flashSaleCache, producer, idGen, logger.Logger)
	query := application.NewFlashsaleQuery(repo)
	service := application.NewFlashsaleService(manager, query)

	// 6. 资源回收
	cleanup := func() {
		bootLog.Info("performing graceful shutdown...")
		clientCleanup()
		producer.Close()
		redisCache.Close()
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}

	return &AppContext{
		Config:     c,
		AppService: service,
		Clients:    clients,
		Metrics:    m,
		Limiter:    rateLimiter,
	}, cleanup, nil
}

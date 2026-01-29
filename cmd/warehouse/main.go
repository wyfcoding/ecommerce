package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/goapi/warehouse/v1"
	"github.com/wyfcoding/ecommerce/internal/warehouse/application"
	"github.com/wyfcoding/ecommerce/internal/warehouse/infrastructure/messaging"
	"github.com/wyfcoding/ecommerce/internal/warehouse/infrastructure/persistence"
	warehousegrpc "github.com/wyfcoding/ecommerce/internal/warehouse/interfaces/grpc"
	warehousehttp "github.com/wyfcoding/ecommerce/internal/warehouse/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/idempotency"
	"github.com/wyfcoding/pkg/limiter"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/messagequeue/outbox"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
)

// BootstrapName 服务唯一标识
const BootstrapName = "warehouse"

// IdempotencyPrefix 幂等性 Redis 键前缀
const IdempotencyPrefix = "warehouse:idem"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

// AppContext 应用上下文 (包含对外服务实例与依赖)
type AppContext struct {
	Config      *Config
	Warehouse   *application.WarehouseService
	Clients     *ServiceClients
	Handler     *warehousehttp.Handler
	Metrics     *metrics.Metrics
	Limiter     limiter.Limiter
	Idempotency idempotency.Manager
}

// ServiceClients 下游微服务客户端集合
type ServiceClients struct {
}

func main() {
	if err := app.NewBuilder[*Config, *AppContext](BootstrapName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGinMiddleware(
			middleware.CORS(),
			middleware.TimeoutMiddleware(30*time.Second),
		).
		Build().
		Run(); err != nil {
		slog.Error("service bootstrap failed", "error", err)
	}
}

func registerGRPC(s *grpc.Server, ctx *AppContext) {
	pb.RegisterWarehouseServiceServer(s, warehousegrpc.NewServer(ctx.Warehouse))
}

func registerGin(e *gin.Engine, ctx *AppContext) {
	if ctx.Config.Server.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	sys := e.Group("/sys")
	{
		sys.GET("/health", func(c *gin.Context) {
			response.SuccessWithRawData(c, gin.H{
				"status":    "UP",
				"service":   BootstrapName,
				"timestamp": time.Now().Unix(),
			})
		})
	}

	if ctx.Config.Metrics.Enabled {
		e.GET(ctx.Config.Metrics.Path, gin.WrapH(ctx.Metrics.Handler()))
	}

	e.Use(middleware.RateLimitWithLimiter(ctx.Limiter))

	api := e.Group("/api/v1")
	{
		ctx.Handler.RegisterRoutes(api)
	}
}

func initService(cfg *Config, m *metrics.Metrics) (*AppContext, func(), error) {
	c := cfg
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default()

	configpkg.PrintWithMask(c)

	db, err := database.NewDB(c.Data.Database, c.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("database init error: %w", err)
	}

	redisCache, err := cache.NewRedisCache(&c.Data.Redis, c.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("redis init error: %w", err)
	}

	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), c.RateLimit.Rate, c.RateLimit.Burst)
	idemManager := idempotency.NewRedisManager(redisCache.GetClient(), IdempotencyPrefix)

	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, m, c.CircuitBreaker, clients)
	if err != nil {
		return nil, nil, fmt.Errorf("grpc clients init error: %w", err)
	}

	// Outbox initialization
	outboxMgr := outbox.NewManager(db.RawDB(), logger.Logger)
	outboxPublisher := messaging.NewOutboxPublisher(outboxMgr)

	// 这里通常需要一个 MQ Pusher，暂设空或从配置初始化
	outboxProcessor := outbox.NewProcessor(outboxMgr, func(ctx context.Context, topic, key string, payload []byte) error {
		// 实际应投递到 Kafka
		bootLog.Info("outbox msg produced (dummy)", "topic", topic, "key", key)
		return nil
	}, 100, 5*time.Second)
	outboxProcessor.Start()

	warehouseRepo := persistence.NewWarehouseRepository(db.RawDB())

	cmdService := application.NewWarehouseCommandService(warehouseRepo, outboxPublisher, logger.Logger)
	queryService := application.NewWarehouseQueryService(warehouseRepo)
	warehouseService := application.NewWarehouseService(cmdService, queryService)

	handler := warehousehttp.NewHandler(warehouseService, logger.Logger)

	cleanup := func() {
		bootLog.Info("shutting down, releasing resources...")
		outboxProcessor.Stop()
		clientCleanup()
		if redisCache != nil {
			redisCache.Close()
		}
		if sqlDB, err := db.RawDB().DB(); err == nil && sqlDB != nil {
			sqlDB.Close()
		}
	}

	return &AppContext{
		Config:      c,
		Warehouse:   warehouseService,
		Clients:     clients,
		Handler:     handler,
		Metrics:     m,
		Limiter:     rateLimiter,
		Idempotency: idemManager,
	}, cleanup, nil
}

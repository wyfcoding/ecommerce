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

	pb "github.com/wyfcoding/ecommerce/go-api/logistics/v1"
	"github.com/wyfcoding/ecommerce/internal/logistics/application"
	"github.com/wyfcoding/ecommerce/internal/logistics/domain"
	logisticssearch "github.com/wyfcoding/ecommerce/internal/logistics/infrastructure/persistence/elasticsearch"
	logisticsmysql "github.com/wyfcoding/ecommerce/internal/logistics/infrastructure/persistence/mysql"
	logisticsredis "github.com/wyfcoding/ecommerce/internal/logistics/infrastructure/persistence/redis"
	logisticsconsumer "github.com/wyfcoding/ecommerce/internal/logistics/interfaces/consumer"
	logisticsgrpc "github.com/wyfcoding/ecommerce/internal/logistics/interfaces/grpc"
	logisticshttp "github.com/wyfcoding/ecommerce/internal/logistics/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/idempotency"
	"github.com/wyfcoding/pkg/limiter"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/messagequeue/outbox"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
	"github.com/wyfcoding/pkg/search"
)

// BootstrapName 服务唯一标识
const BootstrapName = "logistics"

// IdempotencyPrefix 幂等性 Redis 键前缀
const IdempotencyPrefix = "logistics:idem"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
	Search           struct {
		LogisticsIndex string `mapstructure:"logistics_index" toml:"logistics_index"`
	} `mapstructure:"search" toml:"search"`
}

// AppContext 应用上下文 (包含对外服务实例与依赖)
type AppContext struct {
	Config      *Config
	Cmd         *application.LogisticsCommandService
	Query       *application.LogisticsQueryService
	Handler     *logisticshttp.Handler
	Metrics     *metrics.Metrics
	Limiter     limiter.Limiter
	Idempotency idempotency.Manager
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
	pb.RegisterLogisticsServiceServer(s, logisticsgrpc.NewServer(ctx.Cmd, ctx.Query))
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

	bootLog.Info("initializing elasticsearch client...")
	esClient, err := search.NewClient(&search.Config{
		ServiceName:         BootstrapName,
		ElasticsearchConfig: c.Data.Elasticsearch,
		BreakerConfig:       c.CircuitBreaker,
		SlowThreshold:       800 * time.Millisecond,
		MaxRetries:          3,
	}, logger, m)
	if err != nil {
		if redisCache != nil {
			redisCache.Close()
		}
		return nil, nil, fmt.Errorf("elasticsearch init error: %w", err)
	}

	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), c.RateLimit.Rate, c.RateLimit.Burst)
	idemManager := idempotency.NewRedisManager(redisCache.GetClient(), IdempotencyPrefix)

	// Kafka & Outbox initialization
	bootLog.Info("initializing kafka producer and outbox...")
	producer := kafka.NewProducer(&c.MessageQueue.Kafka, logger, m)
	masterDB := db.RawDB()
	if err := masterDB.AutoMigrate(&outbox.Message{}); err != nil {
		return nil, nil, fmt.Errorf("failed to migrate outbox table: %w", err)
	}
	outboxMgr := outbox.NewManager(masterDB, logger.Logger)
	outboxProc := outbox.NewProcessor(outboxMgr, func(ctx context.Context, topic, key string, payload []byte) error {
		return producer.PublishToTopic(ctx, topic, []byte(key), payload)
	}, 100, 5*time.Second)
	outboxProc.Start()

	// Infrastructure
	logisticsRepo := logisticsmysql.NewLogisticsRepository(db.RawDB())
	logisticsReadRepo := logisticsredis.NewLogisticsReadRepository(redisCache.GetClient(), c.Cache.DefaultExpiration)
	logisticsSearchRepo := logisticssearch.NewLogisticsSearchRepository(esClient, c.Search.LogisticsIndex)
	publisher := outbox.NewPublisher(outboxMgr)

	// Application
	cmdService := application.NewLogisticsCommandService(logisticsRepo, publisher, logger.Logger)
	queryService := application.NewLogisticsQueryService(logisticsRepo, logisticsReadRepo, logisticsSearchRepo, logger.Logger)

	// Projection consumers
	projectionService := application.NewLogisticsProjectionService(logisticsRepo, logisticsReadRepo, logisticsSearchRepo, logger.Logger)
	projectionHandler := logisticsconsumer.NewLogisticsProjectionHandler(projectionService, logger.Logger)
	projectionTopics := []string{
		domain.LogisticsCreatedEventType,
		domain.LogisticsStatusUpdatedEventType,
		domain.LogisticsTraceAddedEventType,
		domain.LogisticsRiderAssignedEventType,
	}
	projectionConsumers := make([]*kafka.Consumer, 0, len(projectionTopics))
	for _, topic := range projectionTopics {
		consumerCfg := c.MessageQueue.Kafka
		consumerCfg.Topic = topic
		consumerCfg.GroupID = BootstrapName + "-projection-group"
		projectionConsumer := kafka.NewConsumer(&consumerCfg, logger, m)
		projectionConsumer.Start(context.Background(), 3, projectionHandler.Handle)
		projectionConsumers = append(projectionConsumers, projectionConsumer)
	}

	handler := logisticshttp.NewHandler(cmdService, queryService, logger.Logger)

	cleanup := func() {
		bootLog.Info("shutting down, releasing resources...")
		for _, c := range projectionConsumers {
			if c != nil {
				c.Close()
			}
		}
		outboxProc.Stop()
		if producer != nil {
			producer.Close()
		}
		if redisCache != nil {
			redisCache.Close()
		}
		if sqlDB, err := db.RawDB().DB(); err == nil && sqlDB != nil {
			sqlDB.Close()
		}
	}

	return &AppContext{
		Config:      c,
		Cmd:         cmdService,
		Query:       queryService,
		Handler:     handler,
		Metrics:     m,
		Limiter:     rateLimiter,
		Idempotency: idemManager,
	}, cleanup, nil
}

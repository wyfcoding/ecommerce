package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/go-api/cart/v1"
	"github.com/wyfcoding/ecommerce/internal/cart/application"
	"github.com/wyfcoding/ecommerce/internal/cart/domain"
	cartsearch "github.com/wyfcoding/ecommerce/internal/cart/infrastructure/persistence/elasticsearch"
	cartmysql "github.com/wyfcoding/ecommerce/internal/cart/infrastructure/persistence/mysql"
	cartredis "github.com/wyfcoding/ecommerce/internal/cart/infrastructure/persistence/redis"
	cartconsumer "github.com/wyfcoding/ecommerce/internal/cart/interfaces/consumer"
	cartgrpc "github.com/wyfcoding/ecommerce/internal/cart/interfaces/grpc"
	carthttp "github.com/wyfcoding/ecommerce/internal/cart/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/idempotency"
	"github.com/wyfcoding/pkg/limiter"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/messagequeue/outbox"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
	"github.com/wyfcoding/pkg/response"
	"github.com/wyfcoding/pkg/search"
)

// BootstrapName 服务唯一标识
const BootstrapName = "cart"

// IdempotencyPrefix 幂等性 Redis 键前缀
const IdempotencyPrefix = "cart:idem"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
	Search           struct {
		CartIndex string `mapstructure:"cart_index" toml:"cart_index"`
	} `mapstructure:"search" toml:"search"`
}

// AppContext 应用上下文
type AppContext struct {
	Config      *Config
	Cmd         *application.CartCommandService
	Query       *application.CartQueryService
	Clients     *ServiceClients
	Handler     *carthttp.Handler
	Metrics     *metrics.Metrics
	Limiter     limiter.Limiter
	Idempotency idempotency.Manager
}

// ServiceClients 下游微服务客户端集合
type ServiceClients struct{}

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
	pb.RegisterCartServiceServer(s, cartgrpc.NewServer(ctx.Cmd, ctx.Query))
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
		sys.GET("/ready", func(c *gin.Context) {
			response.SuccessWithRawData(c, gin.H{"status": "READY"})
		})
	}
	if ctx.Config.Metrics.Enabled {
		e.GET(ctx.Config.Metrics.Path, gin.WrapH(ctx.Metrics.Handler()))
	}
	e.Use(middleware.RateLimitWithLimiter(ctx.Limiter))
	api := e.Group("/api/v1")
	{
		api.Use(middleware.JWTAuth(ctx.Config.JWT.Secret))
		ctx.Handler.RegisterRoutes(api)
	}
}

func initService(cfg *Config, m *metrics.Metrics) (*AppContext, func(), error) {
	c := cfg
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default()

	// 打印脱敏配置
	configpkg.PrintWithMask(c)

	// 1. 初始化数据库 (MySQL)
	db, err := database.NewDB(c.Data.Database, c.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, err
	}

	// 2. 初始化缓存
	redisCache, err := cache.NewRedisCache(&c.Data.Redis, c.CircuitBreaker, logger, m)
	if err != nil {
		if sqlDB, err := db.RawDB().DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("redis init error: %w", err)
	}

	// 2.1 初始化 Elasticsearch 客户端 (读模型搜索)
	bootLog.Info("initializing elasticsearch client...")
	esClient, err := search.NewClient(&search.Config{
		ServiceName:         BootstrapName,
		ElasticsearchConfig: c.Data.Elasticsearch,
		BreakerConfig:       c.CircuitBreaker,
		SlowThreshold:       800 * time.Millisecond,
		MaxRetries:          3,
	}, logger, m)
	if err != nil {
		redisCache.Close()
		if sqlDB, err := db.RawDB().DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("elasticsearch init error: %w", err)
	}

	// 3. 初始化治理组件
	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), c.RateLimit.Rate, c.RateLimit.Burst)
	idemManager := idempotency.NewRedisManager(redisCache.GetClient(), IdempotencyPrefix)

	// 3.1 初始化消息队列与 Outbox
	bootLog.Info("initializing kafka producer and outbox...")
	producer := kafka.NewProducer(&c.MessageQueue.Kafka, logger, m)
	if err := db.RawDB().AutoMigrate(&outbox.Message{}); err != nil {
		redisCache.Close()
		if sqlDB, err := db.RawDB().DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("failed to migrate outbox table: %w", err)
	}
	outboxMgr := outbox.NewManager(db.RawDB(), logger.Logger)
	outboxProcessor := outbox.NewProcessor(outboxMgr, func(ctx context.Context, topic, key string, payload []byte) error {
		return producer.PublishToTopic(ctx, topic, []byte(key), payload)
	}, 100, 5*time.Second)
	outboxProcessor.Start()

	// 4. 初始化下游微服务客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, m, c.CircuitBreaker, clients)
	if err != nil {
		outboxProcessor.Stop()
		producer.Close()
		redisCache.Close()
		if sqlDB, err := db.RawDB().DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("grpc clients init error: %w", err)
	}

	// 5. DDD 分层装配
	bootLog.Info("assembling services with full dependency injection...")

	// 5.1 Infrastructure (Persistence)
	cartRepo := cartmysql.NewCartRepository(db.RawDB())
	cartReadRepo := cartredis.NewCartReadRepository(redisCache.GetClient(), c.Cache.DefaultExpiration)
	cartSearchRepo := cartsearch.NewCartSearchRepository(esClient, c.Search.CartIndex)

	// 5.2 Application (Service)
	querySvc := application.NewCartQueryService(cartRepo, cartReadRepo, cartSearchRepo, logger.Logger)
	commandSvc := application.NewCartCommandService(cartRepo, outbox.NewPublisher(outboxMgr), logger.Logger, querySvc)

	// 5.3 Projection Consumers (Cart Events -> Read Model)
	projectionService := application.NewCartProjectionService(cartRepo, cartReadRepo, cartSearchRepo, logger.Logger)
	projectionHandler := cartconsumer.NewCartProjectionHandler(projectionService, logger.Logger)
	projectionTopics := []string{
		domain.CartItemAddedEventType,
		domain.CartItemUpdatedEventType,
		domain.CartItemRemovedEventType,
		domain.CartClearedEventType,
		domain.CartMergedEventType,
		domain.CartCouponAppliedEventType,
		domain.CartCouponRemovedEventType,
	}
	projectionConsumers := make([]*kafka.Consumer, 0, len(projectionTopics))
	for _, topic := range projectionTopics {
		consumerCfg := c.MessageQueue.Kafka
		consumerCfg.Topic = topic
		consumerCfg.GroupID = BootstrapName + "-projection-group"
		consumer := kafka.NewConsumer(&consumerCfg, logger, m)
		consumer.Start(context.Background(), 3, projectionHandler.Handle)
		projectionConsumers = append(projectionConsumers, consumer)
	}

	// 5.4 订单确认事件消费者 (清理购物车)
	orderConsumerCfg := c.MessageQueue.Kafka
	orderConsumerCfg.Topic = "order.confirmed"
	orderConsumerCfg.GroupID = BootstrapName + "-order-confirmed-group"
	orderConsumer := kafka.NewConsumer(&orderConsumerCfg, logger, m)
	orderHandler := cartconsumer.NewOrderConfirmedHandler(commandSvc, idemManager, logger.Logger)
	orderConsumer.Start(context.Background(), 3, orderHandler.Handle)

	// 5.5 Interface (HTTP Handlers)
	handler := carthttp.NewHandler(commandSvc, querySvc, logger.Logger)

	cleanup := func() {
		bootLog.Info("shutting down, releasing resources...")
		if orderConsumer != nil {
			orderConsumer.Close()
		}
		for _, c := range projectionConsumers {
			if c != nil {
				c.Close()
			}
		}
		outboxProcessor.Stop()
		if producer != nil {
			producer.Close()
		}
		clientCleanup()
		if redisCache != nil {
			if err := redisCache.Close(); err != nil {
				bootLog.Error("failed to close redis cache", "error", err)
			}
		}
		if sqlDB, err := db.RawDB().DB(); err == nil && sqlDB != nil {
			if err := sqlDB.Close(); err != nil {
				bootLog.Error("failed to close sql database", "error", err)
			}
		}
	}

	return &AppContext{
		Config: c, Cmd: commandSvc, Query: querySvc, Handler: handler, Metrics: m,
		Limiter: rateLimiter, Idempotency: idemManager,
	}, cleanup, nil
}

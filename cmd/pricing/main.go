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

	dynamicpricingv1 "github.com/wyfcoding/ecommerce/goapi/dynamicpricing/v1"
	pb "github.com/wyfcoding/ecommerce/goapi/pricing/v1"
	"github.com/wyfcoding/ecommerce/internal/pricing/application"
	"github.com/wyfcoding/ecommerce/internal/pricing/domain"
	pricingsearch "github.com/wyfcoding/ecommerce/internal/pricing/infrastructure/persistence/elasticsearch"
	pricingmysql "github.com/wyfcoding/ecommerce/internal/pricing/infrastructure/persistence/mysql"
	pricingredis "github.com/wyfcoding/ecommerce/internal/pricing/infrastructure/persistence/redis"
	pricingconsumer "github.com/wyfcoding/ecommerce/internal/pricing/interfaces/consumer"
	pricinggrpc "github.com/wyfcoding/ecommerce/internal/pricing/interfaces/grpc"
	pricinghttp "github.com/wyfcoding/ecommerce/internal/pricing/interfaces/http"
	marketdatav1 "github.com/wyfcoding/financialtrading/go-api/marketdata/v1"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/grpcclient"
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
const BootstrapName = "pricing"

// IdempotencyPrefix 幂等性 Redis 键前缀
const IdempotencyPrefix = "pricing:idem"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
	Search           struct {
		HistoryIndex string `mapstructure:"history_index" toml:"history_index"`
	} `mapstructure:"search" toml:"search"`
}

// AppContext 应用上下文 (包含对外服务实例与依赖)
type AppContext struct {
	Config      *Config
	Cmd         *application.PricingCommandService
	Query       *application.PricingQueryService
	Clients     *ServiceClients
	Handler     *pricinghttp.Handler
	Metrics     *metrics.Metrics
	Limiter     limiter.Limiter
	Idempotency idempotency.Manager
}

// ServiceClients 下游微服务客户端集合
type ServiceClients struct {
	MarketDataConn     *grpc.ClientConn `service:"marketdata"`
	MarketData         marketdatav1.MarketDataServiceClient
	DynamicPricingConn *grpc.ClientConn `service:"dynamicpricing"`
	DynamicPricing     dynamicpricingv1.DynamicPricingServiceClient
}

func main() {
	// 构建并运行服务
	if err := app.NewBuilder[*Config, *AppContext](BootstrapName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGinMiddleware(
			middleware.CORS(), // 跨域处理
			middleware.TimeoutMiddleware(30*time.Second), // 全局超时
		).
		Build().
		Run(); err != nil {
		slog.Error("service bootstrap failed", "error", err)
	}
}

// registerGRPC 注册 gRPC 服务
func registerGRPC(s *grpc.Server, ctx *AppContext) {
	pb.RegisterPricingServiceServer(s, pricinggrpc.NewServer(ctx.Cmd, ctx.Query))
}

// registerGin 注册 HTTP 路由
func registerGin(e *gin.Engine, ctx *AppContext) {
	// 根据环境设置 Gin 模式
	if ctx.Config.Server.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 系统检查接口
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

	// 指标暴露
	if ctx.Config.Metrics.Enabled {
		e.GET(ctx.Config.Metrics.Path, gin.WrapH(ctx.Metrics.Handler()))
	}

	// 全局限流中间件
	e.Use(middleware.RateLimitWithLimiter(ctx.Limiter))

	// 业务 API 路由 v1
	api := e.Group("/api/v1")
	{
		ctx.Handler.RegisterRoutes(api)
	}
}

// initService 初始化服务依赖 (数据库、缓存、客户端、领域层)
func initService(cfg *Config, m *metrics.Metrics) (*AppContext, func(), error) {
	c := cfg
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default() // 获取全局 Logger

	// 打印脱敏配置
	configpkg.PrintWithMask(c)

	// 1. 初始化数据库 (MySQL)
	db, err := database.NewDB(c.Data.Database, c.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("database init error: %w", err)
	}

	// 2. 初始化缓存 (Redis)
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

	// 3. 初始化治理组件 (限流器、幂等管理器)
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
	// 显式转换 gRPC 客户端 (Cross-Project Bridge)
	if clients.MarketDataConn != nil {
		clients.MarketData = marketdatav1.NewMarketDataServiceClient(clients.MarketDataConn)
	}
	if clients.DynamicPricingConn != nil {
		clients.DynamicPricing = dynamicpricingv1.NewDynamicPricingServiceClient(clients.DynamicPricingConn)
	}

	// 6. DDD 分层装配
	bootLog.Info("assembling services with full dependency injection...")

	// 5.1 Infrastructure (Persistence)
	pricingRepo := pricingmysql.NewPricingRepository(db.RawDB())
	ruleReadRepo := pricingredis.NewPricingRuleReadRepository(redisCache.GetClient(), c.Cache.DefaultExpiration)
	historySearchRepo := pricingsearch.NewPriceHistorySearchRepository(esClient, c.Search.HistoryIndex)

	// 6.2 Application (Service)
	querySvc := application.NewPricingQueryService(pricingRepo, ruleReadRepo, historySearchRepo)
	if clients.MarketData != nil {
		querySvc.SetMarketDataClient(clients.MarketData)
	}
	if clients.DynamicPricing != nil {
		querySvc.SetDynamicPricingClient(clients.DynamicPricing)
	}
	commandSvc := application.NewPricingCommandService(pricingRepo, outbox.NewPublisher(outboxMgr), logger.Logger)

	// 5.3 Projection Consumers (Pricing Events -> Read Model)
	projectionService := application.NewPricingProjectionService(pricingRepo, ruleReadRepo, historySearchRepo, logger.Logger)
	projectionHandler := pricingconsumer.NewPricingProjectionHandler(projectionService, logger.Logger)
	projectionTopics := []string{
		domain.PricingRuleUpdatedEventType,
		domain.PriceHistoryRecordedEventType,
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

	// 5.4 Interface (HTTP Handlers)
	handler := pricinghttp.NewHandler(commandSvc, querySvc, logger.Logger)

	// 定义资源清理函数
	cleanup := func() {
		bootLog.Info("shutting down, releasing resources...")
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

	// 返回应用上下文与清理函数
	return &AppContext{
		Config:      c,
		Cmd:         commandSvc,
		Query:       querySvc,
		Clients:     clients,
		Handler:     handler,
		Metrics:     m,
		Limiter:     rateLimiter,
		Idempotency: idempotency.Manager(idemManager),
	}, cleanup, nil
}

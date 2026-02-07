package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/goapi/aftersales/v1"
	orderv1 "github.com/wyfcoding/ecommerce/goapi/order/v1"
	paymentv1 "github.com/wyfcoding/ecommerce/goapi/payment/v1"
	"github.com/wyfcoding/ecommerce/internal/aftersales/application"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain"
	aftersalessearch "github.com/wyfcoding/ecommerce/internal/aftersales/infrastructure/persistence/elasticsearch"
	aftersalesmysql "github.com/wyfcoding/ecommerce/internal/aftersales/infrastructure/persistence/mysql"
	aftersalesredis "github.com/wyfcoding/ecommerce/internal/aftersales/infrastructure/persistence/redis"
	aftersalesconsumer "github.com/wyfcoding/ecommerce/internal/aftersales/interfaces/consumer"
	aftersalesgrpc "github.com/wyfcoding/ecommerce/internal/aftersales/interfaces/grpc"
	aftersaleshttp "github.com/wyfcoding/ecommerce/internal/aftersales/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/idempotency"
	"github.com/wyfcoding/pkg/idgen"
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
const BootstrapName = "aftersales"

// IdempotencyPrefix 幂等性 Redis 键前缀
const IdempotencyPrefix = "aftersales:idem"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
	Search           struct {
		AfterSalesIndex           string `mapstructure:"after_sales_index" toml:"after_sales_index"`
		SupportTicketIndex        string `mapstructure:"support_ticket_index" toml:"support_ticket_index"`
		SupportTicketMessageIndex string `mapstructure:"support_ticket_message_index" toml:"support_ticket_message_index"`
	} `mapstructure:"search" toml:"search"`
}

// AppContext 应用上下文 (包含对外服务实例与依赖)
type AppContext struct {
	Config      *Config
	Cmd         *application.AfterSalesCommandService
	Query       *application.AfterSalesQueryService
	Clients     *ServiceClients
	Handler     *aftersaleshttp.Handler
	Metrics     *metrics.Metrics
	Limiter     limiter.Limiter
	Idempotency idempotency.Manager
}

// ServiceClients 下游微服务客户端集合
type ServiceClients struct {
	Order   *grpc.ClientConn `service:"order"`
	Payment *grpc.ClientConn `service:"payment"`
}

func main() {
	// 构建并运行服务
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

// registerGRPC 注册 gRPC 服务
func registerGRPC(s *grpc.Server, ctx *AppContext) {
	orderClient := orderv1.NewOrderServiceClient(ctx.Clients.Order)
	pb.RegisterAftersalesServiceServer(s, aftersalesgrpc.NewServer(ctx.Cmd, ctx.Query, orderClient))
}

// registerGin 注册 HTTP 路由
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
		ctx.Handler.RegisterRoutes(api)
	}
}

// initService 初始化服务依赖 (数据库、缓存、客户端、领域层)
func initService(cfg *Config, m *metrics.Metrics) (*AppContext, func(), error) {
	c := cfg
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default()

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

	// 3. 初始化治理组件 (限流器、幂等管理器、ID 生成器)
	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), c.RateLimit.Rate, c.RateLimit.Burst)
	idemManager := idempotency.NewRedisManager(redisCache.GetClient(), IdempotencyPrefix)
	idGenerator, err := idgen.NewGenerator(c.Snowflake)
	if err != nil {
		redisCache.Close()
		if sqlDB, err := db.RawDB().DB(); err == nil {
			if cerr := sqlDB.Close(); cerr != nil {
				bootLog.Error("failed to close sql database", "error", cerr)
			}
		}
		return nil, nil, fmt.Errorf("id generator init error: %w", err)
	}

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
	repo := aftersalesmysql.NewAfterSalesRepository(db.RawDB())
	readRepo := aftersalesredis.NewAfterSalesReadRepository(redisCache.GetClient(), c.Cache.DefaultExpiration)
	supportTicketReadRepo := aftersalesredis.NewSupportTicketReadRepository(redisCache.GetClient(), c.Cache.DefaultExpiration)
	configReadRepo := aftersalesredis.NewConfigReadRepository(redisCache.GetClient(), c.Cache.DefaultExpiration)
	afterSalesSearchRepo := aftersalessearch.NewAfterSalesSearchRepository(esClient, c.Search.AfterSalesIndex)
	supportTicketSearchRepo := aftersalessearch.NewSupportTicketSearchRepository(esClient, c.Search.SupportTicketIndex)
	supportTicketMessageSearchRepo := aftersalessearch.NewSupportTicketMessageSearchRepository(esClient, c.Search.SupportTicketMessageIndex)

	// 5.2 Application (Service)
	orderClient := orderv1.NewOrderServiceClient(clients.Order)
	paymentClient := paymentv1.NewPaymentServiceClient(clients.Payment)

	dtmAddr := c.Services["dtm"].GRPCAddr
	if dtmAddr == "" {
		dtmAddr = "dtm:36789"
	}
	orderSvcURL := c.Services["order"].GRPCAddr
	if orderSvcURL == "" {
		orderSvcURL = "order:50051"
	}
	paymentSvcURL := c.Services["payment"].GRPCAddr
	if paymentSvcURL == "" {
		paymentSvcURL = "payment:50051"
	}
	aftersalesURL := c.Services["aftersales"].GRPCAddr
	if aftersalesURL == "" {
		aftersalesURL = "aftersales:50051"
	}

	commandSvc := application.NewAfterSalesCommandService(
		repo,
		outbox.NewPublisher(outboxMgr),
		idGenerator,
		logger.Logger,
		orderClient,
		paymentClient,
		dtmAddr,
		orderSvcURL,
		paymentSvcURL,
		aftersalesURL,
	)
	querySvc := application.NewAfterSalesQueryService(
		repo,
		readRepo,
		afterSalesSearchRepo,
		supportTicketReadRepo,
		supportTicketSearchRepo,
		supportTicketMessageSearchRepo,
		configReadRepo,
	)

	// 5.3 Projection Consumers (Aftersales Events -> Read Model)
	projectionService := application.NewAfterSalesProjectionService(
		repo,
		readRepo,
		afterSalesSearchRepo,
		supportTicketReadRepo,
		supportTicketSearchRepo,
		supportTicketMessageSearchRepo,
		configReadRepo,
		logger.Logger,
	)
	projectionHandler := aftersalesconsumer.NewAfterSalesProjectionHandler(projectionService, logger.Logger)
	projectionTopics := []string{
		domain.AfterSalesCreatedEventType,
		domain.AfterSalesStatusUpdatedEventType,
		domain.AfterSalesSupportTicketCreatedType,
		domain.AfterSalesSupportTicketUpdatedType,
		domain.AfterSalesSupportTicketMessageType,
		domain.AfterSalesConfigUpdatedEventType,
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
	handler := aftersaleshttp.NewHandler(commandSvc, querySvc, logger.Logger)

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

	return &AppContext{
		Config:      c,
		Cmd:         commandSvc,
		Query:       querySvc,
		Clients:     clients,
		Handler:     handler,
		Metrics:     m,
		Limiter:     rateLimiter,
		Idempotency: idemManager,
	}, cleanup, nil
}

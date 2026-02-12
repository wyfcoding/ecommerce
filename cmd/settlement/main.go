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

	pb "github.com/wyfcoding/ecommerce/go-api/settlement/v1"
	"github.com/wyfcoding/ecommerce/internal/settlement/application"
	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
	settlementsearch "github.com/wyfcoding/ecommerce/internal/settlement/infrastructure/persistence/elasticsearch"
	settlementmysql "github.com/wyfcoding/ecommerce/internal/settlement/infrastructure/persistence/mysql"
	settlementredis "github.com/wyfcoding/ecommerce/internal/settlement/infrastructure/persistence/redis"
	consumer "github.com/wyfcoding/ecommerce/internal/settlement/interfaces/consumer"
	settlementgrpc "github.com/wyfcoding/ecommerce/internal/settlement/interfaces/grpc"
	settlementhttp "github.com/wyfcoding/ecommerce/internal/settlement/interfaces/http"
	accountv1 "github.com/wyfcoding/financialtrading/go-api/account/v1"
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
const BootstrapName = "settlement"

// IdempotencyPrefix 幂等性 Redis 键前缀
const IdempotencyPrefix = "settlement:idem"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
	Search           struct {
		SettlementIndex string `mapstructure:"settlement_index" toml:"settlement_index"`
	} `mapstructure:"search" toml:"search"`
}

// AppContext 应用上下文 (包含对外服务实例与依赖)
type AppContext struct {
	Config      *Config
	Cmd         *application.SettlementCommandService
	Query       *application.SettlementQueryService
	Clients     *ServiceClients
	Handler     *settlementhttp.Handler
	Metrics     *metrics.Metrics
	Limiter     limiter.Limiter
	Idempotency idempotency.Manager
}

// ServiceClients 下游微服务客户端集合
type ServiceClients struct {
	AccountConn *grpc.ClientConn `service:"account"`
	Account     accountv1.AccountServiceClient
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
	pb.RegisterSettlementServiceServer(s, settlementgrpc.NewServer(ctx.Cmd, ctx.Query))
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
	outboxProc := outbox.NewProcessor(outboxMgr, func(ctx context.Context, topic, key string, payload []byte) error {
		return producer.PublishToTopic(ctx, topic, []byte(key), payload)
	}, 100, 5*time.Second)
	outboxProc.Start()

	// 4. 初始化下游微服务客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, m, c.CircuitBreaker, clients)
	if err != nil {
		if outboxProc != nil {
			outboxProc.Stop()
		}
		if producer != nil {
			producer.Close()
		}
		redisCache.Close()
		if sqlDB, err := db.RawDB().DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("grpc clients init error: %w", err)
	}
	// 显式转换 gRPC 客户端 (Cross-Project Bridge)
	if clients.AccountConn != nil {
		clients.Account = accountv1.NewAccountServiceClient(clients.AccountConn)
	}

	// 5. DDD 分层装配
	bootLog.Info("assembling services with full dependency injection...")

	// 5.1 Infrastructure (Persistence & Messaging)
	settlementRepo := settlementmysql.NewSettlementRepository(db.RawDB())
	ledgerRepo := settlementmysql.NewLedgerRepository(db.RawDB())
	ledgerService := domain.NewLedgerService(ledgerRepo)

	settlementReadRepo := settlementredis.NewSettlementReadRepository(redisCache.GetClient(), c.Cache.DefaultExpiration)
	settlementSearchRepo := settlementsearch.NewSettlementSearchRepository(esClient, c.Search.SettlementIndex)

	// 初始化可靠消息发布者 (Outbox)
	publisher := outbox.NewPublisher(outboxMgr)

	// 5.2 Application (Service)
	cmdService := application.NewSettlementCommandService(settlementRepo, ledgerService, publisher, logger.Logger)
	if clients.Account != nil {
		cmdService.SetRemoteAccountClient(clients.Account)
	}
	queryService := application.NewSettlementQueryService(settlementRepo, settlementReadRepo, settlementSearchRepo)

	// 5.3 Projection Consumers (Settlement Events -> Read Model)
	projectionService := application.NewSettlementProjectionService(settlementRepo, settlementReadRepo, settlementSearchRepo, logger.Logger)
	projectionHandler := consumer.NewSettlementProjectionHandler(projectionService, logger.Logger)
	projectionTopics := []string{
		"settlement.created",
		"settlement.processed",
		"settlement.completed",
		"settlement.failed",
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

	// 5.4 Event Handlers (Kafka Consumer)
	bootLog.Info("initializing kafka consumer for payment events...")
	paymentHandler := consumer.NewPaymentHandler(cmdService, logger.Logger)

	// 配置消费者组
	paymentConsumerCfg := c.MessageQueue.Kafka
	paymentConsumerCfg.Topic = "payment.captured"
	paymentConsumerCfg.GroupID = BootstrapName + "-group"

	paymentConsumer := kafka.NewConsumer(&paymentConsumerCfg, logger, m)
	paymentConsumer.Start(context.Background(), 3, paymentHandler.HandlePaymentCaptured)

	// 5.5 Interface (HTTP Handlers)
	handler := settlementhttp.NewHandler(cmdService, queryService, logger.Logger)

	// 定义资源清理函数
	cleanup := func() {
		bootLog.Info("shutting down, releasing resources...")
		for _, c := range projectionConsumers {
			if c != nil {
				c.Close()
			}
		}
		if paymentConsumer != nil {
			paymentConsumer.Close()
		}
		if outboxProc != nil {
			outboxProc.Stop()
		}
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
		Cmd:         cmdService,
		Query:       queryService,
		Clients:     clients,
		Handler:     handler,
		Metrics:     m,
		Limiter:     rateLimiter,
		Idempotency: idempotency.Manager(idemManager),
	}, cleanup, nil
}

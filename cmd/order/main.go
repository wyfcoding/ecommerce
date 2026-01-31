package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	inventoryv1 "github.com/wyfcoding/ecommerce/goapi/inventory/v1"
	pb "github.com/wyfcoding/ecommerce/goapi/order/v1"
	paymentv1 "github.com/wyfcoding/ecommerce/goapi/payment/v1"
	productv1 "github.com/wyfcoding/ecommerce/goapi/product/v1"
	"github.com/wyfcoding/ecommerce/internal/order/application"
	"github.com/wyfcoding/ecommerce/internal/order/infrastructure/messaging"
	"github.com/wyfcoding/ecommerce/internal/order/infrastructure/persistence"
	"github.com/wyfcoding/ecommerce/internal/order/interfaces/consumer"
	ordergrpc "github.com/wyfcoding/ecommerce/internal/order/interfaces/grpc"
	orderhttp "github.com/wyfcoding/ecommerce/internal/order/interfaces/http"
	positionv1 "github.com/wyfcoding/financialtrading/go-api/position/v1"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database/sharding"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/idempotency"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/limiter"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/messagequeue/outbox"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
	"github.com/wyfcoding/pkg/security/risk"
)

// BootstrapName 服务唯一标识
const BootstrapName = "order"

// IdempotencyPrefix 幂等性 Redis 键前缀
const IdempotencyPrefix = "order:idem"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

// AppContext 应用上下文 (包含对外服务实例与依赖)
type AppContext struct {
	Config      *Config
	Cmd         *application.OrderCommandService
	Query       *application.OrderQueryService
	Clients     *ServiceClients
	Handler     *orderhttp.Handler
	Metrics     *metrics.Metrics
	Limiter     limiter.Limiter
	Idempotency idempotency.Manager
}

// ServiceClients 下游微服务客户端集合
type ServiceClients struct {
	Warehouse *grpc.ClientConn `service:"warehouse"`
	Inventory *grpc.ClientConn `service:"inventory"`
	Payment   *grpc.ClientConn `service:"payment"`
	Product   *grpc.ClientConn `service:"product"`
	Position  *grpc.ClientConn `service:"position"` // FinancialTrading Position Service
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
	pb.RegisterOrderServiceServer(s, ordergrpc.NewServer(ctx.Cmd, ctx.Query))
}

// registerGin 注册 HTTP 路由
func registerGin(e *gin.Engine, ctx *AppContext) {
	// 根据环境设置 Gin 模式
	if ctx.Config.Server.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 全局限流中间件 (使用从 AppContext 获取的 Redis 限流器)
	e.Use(middleware.RateLimitWithLimiter(ctx.Limiter))

	// 业务 API 路由 v1
	api := e.Group("/api/v1")
	{
		// 鉴权 (订单接口通常需要)
		api.Use(middleware.JWTAuth(ctx.Config.JWT.Secret))
		// 幂等 (订单提交、支付等必选)
		api.Use(middleware.IdempotencyMiddleware(ctx.Idempotency, 24*time.Hour))

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

	// 1. 初始化分片数据库 (MySQL Sharding)
	bootLog.Info("initializing sharding database manager...")
	shardingManager, err := sharding.NewManager(c.Data.Shards, c.CircuitBreaker, logger, m)
	if err != nil {
		// 容错：如果 Shards 为空，退回到单库
		shardingManager, err = sharding.NewManager([]configpkg.DatabaseConfig{c.Data.Database}, c.CircuitBreaker, logger, m)
		if err != nil {
			return nil, nil, fmt.Errorf("sharding database init error: %w", err)
		}
	}

	// 2. 初始化缓存 (Redis)
	bootLog.Info("initializing redis cache...")
	redisCache, err := cache.NewRedisCache(&c.Data.Redis, c.CircuitBreaker, logger, m)
	if err != nil {
		shardingManager.Close()
		return nil, nil, fmt.Errorf("redis init error: %w", err)
	}

	// 3. 初始化消息队列 (Kafka Producer)
	bootLog.Info("initializing kafka producer...")
	producer := kafka.NewProducer(&c.MessageQueue.Kafka, logger, m)

	// --- 3.1 分片感知型 Outbox 初始化 (顶级架构增强) ---
	allDBs := shardingManager.GetAllDBs()
	outboxProcessors := make([]*outbox.Processor, 0, len(allDBs))
	defaultOutboxMgr := outbox.NewManager(shardingManager.GetDB(0), logger.Logger)

	for i, dbNode := range allDBs {
		bootLog.Info("syncing outbox schema and starting processor for shard", "shard_index", i)
		if err := dbNode.AutoMigrate(&outbox.Message{}); err != nil {
			return nil, nil, fmt.Errorf("failed to migrate outbox table on shard %d: %w", i, err)
		}
		shardMgr := outbox.NewManager(dbNode, logger.Logger)
		proc := outbox.NewProcessor(shardMgr, func(ctx context.Context, topic, key string, payload []byte) error {
			return producer.PublishToTopic(ctx, topic, []byte(key), payload)
		}, 100, 5*time.Second)
		proc.Start()
		outboxProcessors = append(outboxProcessors, proc)
	}

	// 4. 初始化治理组件 (限流器、幂等管理器、ID 生成器、风控引擎)
	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), c.RateLimit.Rate, c.RateLimit.Burst)
	idemManager := idempotency.NewRedisManager(redisCache.GetClient(), IdempotencyPrefix)
	riskEvaluator, err := risk.NewDynamicRiskEngine(logger.Logger)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize risk evaluator: %v", err))
	}
	idGenerator, err := idgen.NewGenerator(c.Snowflake)
	if err != nil {
		return nil, nil, fmt.Errorf("idgen init error: %w", err)
	}

	// 5. 初始化下游微服务客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, m, c.CircuitBreaker, clients)
	if err != nil {
		for _, p := range outboxProcessors {
			p.Stop()
		}
		producer.Close()
		redisCache.Close()
		shardingManager.Close()
		return nil, nil, fmt.Errorf("grpc clients init error: %w", err)
	}

	// 6. DDD 分层装配
	bootLog.Info("assembling services with full dependency injection...")

	// 6.1 Infrastructure (Persistence)
	orderRepo := persistence.NewOrderRepository(shardingManager)

	// 6.2 Application (Service)
	warehouseAddr := c.Services["warehouse"].GRPCAddr
	dtmAddr := c.Services["dtm"].GRPCAddr
	if dtmAddr == "" {
		dtmAddr = "dtm:36789"
	}
	orderSvcAddr := c.Services["order"].GRPCAddr
	if orderSvcAddr == "" {
		orderSvcAddr = "order:50051"
	}

	orderCommandService := application.NewOrderCommandService(
		orderRepo,
		idGenerator,
		messaging.NewOutboxPublisher(defaultOutboxMgr),
		logger.Logger,
		dtmAddr,
		warehouseAddr,
		m,
		riskEvaluator,
	)
	orderCommandService.SetSvcURL(orderSvcAddr)

	// 注入 gRPC 客户端 (Internal Service Interaction)
	orderCommandService.SetClients(
		inventoryv1.NewInventoryServiceClient(clients.Inventory),
		paymentv1.NewPaymentServiceClient(clients.Payment),
		positionv1.NewPositionServiceClient(clients.Position),
		productv1.NewProductServiceClient(clients.Product),
	)

	orderQuery := application.NewOrderQueryService(orderRepo, redisCache, logger.Logger, m)

	// --- 6.3 Event Handlers (Kafka Consumer) ---
	bootLog.Info("initializing kafka consumer for flashsale events...")
	flashsaleHandler := consumer.NewFlashsaleHandler(orderCommandService, logger.Logger)
	flashsaleConsumerCfg := c.MessageQueue.Kafka
	flashsaleConsumerCfg.Topic = "flashsale.order"
	flashsaleConsumerCfg.GroupID = BootstrapName + "-flashsale-group"
	flashsaleConsumer := kafka.NewConsumer(&flashsaleConsumerCfg, logger, m)
	flashsaleConsumer.Start(context.Background(), 5, flashsaleHandler.HandleFlashsaleOrder)

	// 6.4 Interface (HTTP Handlers)
	httpHandler := orderhttp.NewHandler(orderCommandService, orderQuery, logger.Logger)

	// 定义资源清理函数
	cleanup := func() {
		bootLog.Info("shutting down, releasing resources...")
		if flashsaleConsumer != nil {
			flashsaleConsumer.Close()
		}
		for _, p := range outboxProcessors {
			p.Stop()
		}
		clientCleanup()
		if producer != nil {
			producer.Close()
		}
		if redisCache != nil {
			redisCache.Close()
		}
		if shardingManager != nil {
			shardingManager.Close()
		}
	}

	// 返回应用上下文与清理函数
	return &AppContext{
		Config:      c,
		Cmd:         orderCommandService,
		Query:       orderQuery,
		Clients:     clients,
		Handler:     httpHandler,
		Metrics:     m,
		Limiter:     rateLimiter,
		Idempotency: idempotency.Manager(idemManager),
	}, cleanup, nil
}

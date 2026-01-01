package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	inventoryv1 "github.com/wyfcoding/ecommerce/goapi/inventory/v1"
	pb "github.com/wyfcoding/ecommerce/goapi/order/v1"
	paymentv1 "github.com/wyfcoding/ecommerce/goapi/payment/v1"
	"github.com/wyfcoding/ecommerce/internal/order/application"
	"github.com/wyfcoding/ecommerce/internal/order/infrastructure/persistence"
	"github.com/wyfcoding/ecommerce/internal/order/interfaces/event"
	ordergrpc "github.com/wyfcoding/ecommerce/internal/order/interfaces/grpc"
	orderhttp "github.com/wyfcoding/ecommerce/internal/order/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases/sharding"
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
	Order       *application.OrderService
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
}

func main() {
	// 构建并运行服务
	if err := app.NewBuilder(BootstrapName).
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
func registerGRPC(s *grpc.Server, svc any) {
	ctx := svc.(*AppContext)
	pb.RegisterOrderServiceServer(s, ordergrpc.NewServer(ctx.Order))
}

// registerGin 注册 HTTP 路由
func registerGin(e *gin.Engine, svc any) {
	ctx := svc.(*AppContext)

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
		// 鉴权 (订单接口通常需要)
		api.Use(middleware.JWTAuth(ctx.Config.JWT.Secret))
		// 幂等 (订单提交、支付等必选)
		api.Use(middleware.IdempotencyMiddleware(ctx.Idempotency, 24*time.Hour))

		ctx.Handler.RegisterRoutes(api)
	}
}

// initService 初始化服务依赖 (数据库、缓存、客户端、领域层)
func initService(cfg any, m *metrics.Metrics) (any, func(), error) {
	c := cfg.(*Config)
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default() // 获取全局 Logger

	// 打印脱敏配置
	configpkg.PrintWithMask(c)

	// 1. 初始化分片数据库 (MySQL Sharding)
	bootLog.Info("initializing sharding database manager...")
	var (
		shardingManager *sharding.Manager
		err             error
	)
	if len(c.Data.Shards) > 0 {
		shardingManager, err = sharding.NewManager(c.Data.Shards, c.CircuitBreaker, logger, m)
	} else {
		shardingManager, err = sharding.NewManager([]configpkg.DatabaseConfig{c.Data.Database}, c.CircuitBreaker, logger, m)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("sharding database init error: %w", err)
	}

	// 2. 初始化缓存 (Redis)
	bootLog.Info("initializing redis cache...")
	redisCache, err := cache.NewRedisCache(c.Data.Redis, c.CircuitBreaker, logger, m)
	if err != nil {
		shardingManager.Close()
		return nil, nil, fmt.Errorf("redis init error: %w", err)
	}

	// 3. 初始化消息队列 (Kafka Producer)
	bootLog.Info("initializing kafka producer...")
	producer := kafka.NewProducer(c.MessageQueue.Kafka, logger, m)

	// --- 3.1 初始化 Outbox (顶级架构增强：确保 DB 事务与消息发送一致性) ---
	// 在分片环境下，通常每个分片都有自己的 outbox 表。为了简单起见，我们先使用分片 0。
	masterDB := shardingManager.GetDB(0)
	if err := masterDB.AutoMigrate(&outbox.OutboxMessage{}); err != nil {
		return nil, nil, fmt.Errorf("failed to migrate outbox table: %w", err)
	}

	outboxMgr := outbox.NewManager(masterDB, logger.Logger)
	outboxProc := outbox.NewProcessor(outboxMgr, func(ctx context.Context, topic, key string, payload []byte) error {
		// 使用已初始化的 producer 发送，避免重复创建资源
		return producer.PublishToTopic(ctx, topic, []byte(key), payload)
	}, 100, 5*time.Second)
	outboxProc.Start()

	// 4. 初始化治理组件 (限流器、幂等管理器、ID 生成器、风控引擎)
	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), c.RateLimit.Rate, time.Second)
	idemManager := idempotency.NewRedisManager(redisCache.GetClient(), IdempotencyPrefix)
	riskEvaluator := risk.NewDynamicRiskEngine(logger.Logger)

	idGenerator, err := idgen.NewGenerator(c.Snowflake)
	if err != nil {
		return nil, nil, fmt.Errorf("idgen init error: %w", err)
	}

	// 5. 初始化下游微服务客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, m, c.CircuitBreaker, clients)
	if err != nil {
		outboxProc.Stop()
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
	// 注意：NewOrderManager 需要多个依赖项
	warehouseAddr := ""
	if w, ok := c.Services["warehouse"]; ok {
		warehouseAddr = w.GRPCAddr
	}

	orderManager := application.NewOrderManager(
		orderRepo,
		idGenerator,
		producer,
		outboxMgr, // 注入 Outbox Manager
		logger.Logger,
		"localhost:36789", // dtm server address
		warehouseAddr,
		m,
		riskEvaluator,
	)

	// 注入 gRPC 客户端 (Internal Service Interaction)
	if clients.Inventory != nil && clients.Payment != nil {
		orderManager.SetClients(
			inventoryv1.NewInventoryServiceClient(clients.Inventory),
			paymentv1.NewPaymentServiceClient(clients.Payment),
		)
	}
	orderQuery := application.NewOrderQuery(orderRepo)

	orderService := application.NewOrderService(
		orderManager,
		orderQuery,
		logger.Logger,
	)

	// --- 6.3 Event Handlers (Kafka Consumer) ---
	bootLog.Info("initializing kafka consumer for flashsale events...")
	flashsaleHandler := event.NewFlashsaleHandler(orderManager, logger.Logger)

	flashsaleConsumerCfg := c.MessageQueue.Kafka
	flashsaleConsumerCfg.Topic = "flashsale.order"
	flashsaleConsumerCfg.GroupID = BootstrapName + "-flashsale-group"

	flashsaleConsumer := kafka.NewConsumer(flashsaleConsumerCfg, logger, m)
	flashsaleConsumer.Start(context.Background(), 5, flashsaleHandler.HandleFlashsaleOrder)

	// 6.3 Interface (HTTP Handlers)
	handler := orderhttp.NewHandler(orderService, logger.Logger)

	// 定义资源清理函数
	cleanup := func() {
		bootLog.Info("shutting down, releasing resources...")
		if flashsaleConsumer != nil {
			flashsaleConsumer.Close()
		}
		outboxProc.Stop() // 停止 Outbox 处理器
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
		Order:       orderService,
		Clients:     clients,
		Handler:     handler,
		Metrics:     m,
		Limiter:     rateLimiter,
		Idempotency: idemManager,
	}, cleanup, nil
}

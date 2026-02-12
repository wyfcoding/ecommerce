package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
	kafkago "github.com/segmentio/kafka-go"
	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/go-api/inventory/v1"
	orderv1 "github.com/wyfcoding/ecommerce/go-api/order/v1"
	"github.com/wyfcoding/ecommerce/internal/inventory/application"
	inventorysearch "github.com/wyfcoding/ecommerce/internal/inventory/infrastructure/persistence/elasticsearch"
	persistence "github.com/wyfcoding/ecommerce/internal/inventory/infrastructure/persistence/mysql"
	inventoryredis "github.com/wyfcoding/ecommerce/internal/inventory/infrastructure/persistence/redis"
	consumer "github.com/wyfcoding/ecommerce/internal/inventory/interfaces/consumer"
	inventorygrpc "github.com/wyfcoding/ecommerce/internal/inventory/interfaces/grpc"
	inventoryhttp "github.com/wyfcoding/ecommerce/internal/inventory/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database/sharding"
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
const BootstrapName = "inventory"

// IdempotencyPrefix 幂等性 Redis 键前缀
const IdempotencyPrefix = "inventory:idem"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
	Search           struct {
		InventoryIndex string `mapstructure:"inventory_index" toml:"inventory_index"`
	} `mapstructure:"search" toml:"search"`
}

// AppContext 应用上下文 (包含对外服务实例与依赖)
type AppContext struct {
	Config      *Config
	Cmd         *application.InventoryCommandService
	Query       *application.InventoryQueryService
	Clients     *ServiceClients
	Handler     *inventoryhttp.Handler
	Metrics     *metrics.Metrics
	Limiter     limiter.Limiter
	Idempotency idempotency.Manager
	Consumer    *kafka.Consumer
}

// ServiceClients 下游微服务客户端集合
type ServiceClients struct {
	OrderConn *grpc.ClientConn `service:"order"`
	Order     orderv1.OrderServiceClient
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
	pb.RegisterInventoryServiceServer(s, inventorygrpc.NewServer(ctx.Cmd, ctx.Query))
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

	// 1. 初始化分片数据库 (MySQL Sharding)
	bootLog.Info("initializing sharding database manager...")
	var (
		shardingMgr *sharding.Manager
		err         error
	)
	if len(c.Data.Shards) > 0 {
		shardingMgr, err = sharding.NewManager(c.Data.Shards, c.CircuitBreaker, logger, m)
	} else {
		shardingMgr, err = sharding.NewManager([]configpkg.DatabaseConfig{c.Data.Database}, c.CircuitBreaker, logger, m)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("sharding database init error: %w", err)
	}

	// 2. 初始化缓存 (Redis)
	redisCache, err := cache.NewRedisCache(&c.Data.Redis, c.CircuitBreaker, logger, m)
	if err != nil {
		shardingMgr.Close()
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
		shardingMgr.Close()
		redisCache.Close()
		return nil, nil, fmt.Errorf("elasticsearch init error: %w", err)
	}

	// 3. 初始化治理组件 (限流器、幂等管理器)
	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), c.RateLimit.Rate, c.RateLimit.Burst)
	idemManager := idempotency.NewRedisManager(redisCache.GetClient(), IdempotencyPrefix)

	// 4. 初始化下游微服务客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, m, c.CircuitBreaker, clients)
	if err != nil {
		redisCache.Close()
		shardingMgr.Close()
		return nil, nil, fmt.Errorf("grpc clients init error: %w", err)
	}
	// 显式转换 gRPC 客户端
	if clients.OrderConn != nil {
		clients.Order = orderv1.NewOrderServiceClient(clients.OrderConn)
	}

	// 5. DDD 分层装配
	bootLog.Info("assembling services with full dependency injection...")

	// 5.1 Infrastructure (Persistence & Messaging)
	inventoryRepo := persistence.NewInventoryRepository(shardingMgr)
	warehouseRepo := persistence.NewWarehouseRepository(shardingMgr.GetDB(0))
	eventStore := persistence.NewEventStore(shardingMgr)
	inventoryReadRepo := inventoryredis.NewInventoryReadRepository(redisCache.GetClient(), c.Cache.DefaultExpiration)
	inventorySearchRepo := inventorysearch.NewInventorySearchRepository(esClient, c.Search.InventoryIndex)

	// 初始化 Kafka Producer 与 Outbox 处理器
	producer := kafka.NewProducer(&c.MessageQueue.Kafka, logger, m)
	allDBs := shardingMgr.GetAllDBs()
	outboxProcessors := make([]*outbox.Processor, 0, len(allDBs))
	defaultOutboxMgr := outbox.NewManager(shardingMgr.GetDB(0), logger.Logger)

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

	publisher := outbox.NewPublisher(defaultOutboxMgr)

	// 5.2 Application (Service)
	cmdService, err := application.NewInventoryCommandService(inventoryRepo, warehouseRepo, publisher, eventStore, logger.Logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create inventory command service: %w", err)
	}
	if clients.Order != nil {
		cmdService.SetRemoteOrderClient(clients.Order)
	}

	queryService := application.NewInventoryQueryService(inventoryRepo, inventoryReadRepo, inventorySearchRepo, eventStore, logger.Logger)

	// 5.3 Projection Consumers (Inventory Events -> Read Model)
	projectionService := application.NewInventoryProjectionService(inventoryRepo, inventoryReadRepo, inventorySearchRepo, logger.Logger)
	projectionHandler := consumer.NewInventoryProjectionHandler(projectionService, logger.Logger)
	projectionTopics := []string{
		"inventory.stock.locked",
		"inventory.stock.unlocked",
		"inventory.stock.deducted",
		"inventory.stock.added",
		"inventory.stock.warning",
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

	// 5. 启动可靠库存自动释放消费者
	consumer := kafka.NewConsumer(&c.MessageQueue.Kafka, logger, m)
	consumer.Start(context.Background(), 5, func(ctx context.Context, msg kafkago.Message) error {
		if msg.Topic != "order.payment.timeout" {
			return nil
		}
		var event map[string]any
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return err
		}

		orderID := fmt.Sprintf("%v", event["order_id"])
		// --- 幂等保护：防止库存重复释放 ---
		idemKey := fmt.Sprintf("inventory:timeout:%s", orderID)
		isFirst, _, err := idemManager.TryStart(ctx, idemKey, 24*time.Hour)
		if err != nil || !isFirst {
			return err
		}

		if err := cmdService.HandleOrderTimeout(ctx, event); err != nil {
			_ = idemManager.Delete(ctx, idemKey)
			return err
		}

		_ = idemManager.Finish(ctx, idemKey, &idempotency.Response{Body: "OK"}, 24*time.Hour)
		return nil
	})

	// 6. 接口层
	handler := inventoryhttp.NewHandler(cmdService, queryService, logger.Logger)

	// 定义资源清理函数
	cleanup := func() {
		bootLog.Info("shutting down, releasing resources...")
		for _, c := range projectionConsumers {
			if c != nil {
				c.Close()
			}
		}
		for _, p := range outboxProcessors {
			if p != nil {
				p.Stop()
			}
		}
		if producer != nil {
			producer.Close()
		}
		if consumer != nil {
			consumer.Close()
		}
		clientCleanup()
		if redisCache != nil {
			if err := redisCache.Close(); err != nil {
				bootLog.Error("failed to close redis cache", "error", err)
			}
		}
		if shardingMgr != nil {
			if err := shardingMgr.Close(); err != nil {
				bootLog.Error("failed to close sharding manager", "error", err)
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
		Consumer:    consumer,
	}, cleanup, nil
}

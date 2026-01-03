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

	pb "github.com/wyfcoding/ecommerce/goapi/inventory/v1"
	orderv1 "github.com/wyfcoding/ecommerce/goapi/order/v1"
	"github.com/wyfcoding/ecommerce/internal/inventory/application"
	"github.com/wyfcoding/ecommerce/internal/inventory/infrastructure/persistence"
	inventorygrpc "github.com/wyfcoding/ecommerce/internal/inventory/interfaces/grpc"
	inventoryhttp "github.com/wyfcoding/ecommerce/internal/inventory/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases/sharding"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/idempotency"
	"github.com/wyfcoding/pkg/limiter"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
)

// BootstrapName 服务唯一标识
const BootstrapName = "inventory"

// IdempotencyPrefix 幂等性 Redis 键前缀
const IdempotencyPrefix = "inventory:idem"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

// AppContext 应用上下文 (包含对外服务实例与依赖)
type AppContext struct {
	Config      *Config
	Inventory   *application.Inventory
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
	pb.RegisterInventoryServiceServer(s, inventorygrpc.NewServer(ctx.Inventory))
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
	redisCache, err := cache.NewRedisCache(c.Data.Redis, c.CircuitBreaker, logger, m)
	if err != nil {
		shardingMgr.Close()
		return nil, nil, fmt.Errorf("redis init error: %w", err)
	}

	// 3. 初始化治理组件 (限流器、幂等管理器)
	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), c.RateLimit.Rate, time.Second)
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

	// 5.1 Infrastructure (Persistence)
	inventoryRepo := persistence.NewInventoryRepository(shardingMgr)
	warehouseRepo := persistence.NewWarehouseRepository(shardingMgr.GetDB(0))
	manager := application.NewInventoryManager(inventoryRepo, warehouseRepo, logger.Logger)
	if clients.Order != nil {
		manager.SetRemoteOrderClient(clients.Order)
	}
	inventoryService := application.NewInventory(manager, application.NewInventoryQuery(inventoryRepo, warehouseRepo, logger.Logger))

	// 5. 启动可靠库存自动释放消费者
	consumer := kafka.NewConsumer(c.MessageQueue.Kafka, logger, m)
	consumer.Start(context.Background(), 5, func(ctx context.Context, msg kafkago.Message) error {
		if msg.Topic != "order.payment.timeout" {
			return nil
		}
		var event map[string]any
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return err
		}
		return manager.HandleOrderTimeout(ctx, event)
	})

	// 6. 接口层
	handler := inventoryhttp.NewHandler(inventoryService, logger.Logger)

	// 定义资源清理函数
	cleanup := func() {
		bootLog.Info("shutting down, releasing resources...")
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
		Inventory:   inventoryService,
		Clients:     clients,
		Handler:     handler,
		Metrics:     m,
		Limiter:     rateLimiter,
		Idempotency: idempotency.Manager(idemManager),
		Consumer:    consumer,
	}, cleanup, nil
}

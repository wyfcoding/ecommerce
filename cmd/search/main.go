package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	kafkago "github.com/segmentio/kafka-go"
	pb "github.com/wyfcoding/ecommerce/goapi/search/v1"
	"github.com/wyfcoding/ecommerce/internal/search/application"
	"github.com/wyfcoding/ecommerce/internal/search/infrastructure/persistence"
	searchgrpc "github.com/wyfcoding/ecommerce/internal/search/interfaces/grpc"
	searchhttp "github.com/wyfcoding/ecommerce/internal/search/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/idempotency"
	"github.com/wyfcoding/pkg/limiter"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
	pkgsearch "github.com/wyfcoding/pkg/search"
)

// BootstrapName 服务唯一标识
const BootstrapName = "search"

// IdempotencyPrefix 幂等性 Redis 键前缀
const IdempotencyPrefix = "search:idem"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

// AppContext 应用上下文 (包含对外服务实例与依赖)
type AppContext struct {
	Config      *Config
	Search      *application.Search
	Clients     *ServiceClients
	Handler     *searchhttp.Handler
	Metrics     *metrics.Metrics
	Limiter     limiter.Limiter
	Idempotency idempotency.Manager
	Consumer    *kafka.Consumer
}

// ServiceClients 下游微服务客户端集合
type ServiceClients struct {
	// 目前 Search 服务无下游强依赖
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
	pb.RegisterSearchServiceServer(s, searchgrpc.NewServer(ctx.Search, logging.Default().Logger))
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

	// 1. 初始化数据库 (MySQL)
	db, err := databases.NewDB(c.Data.Database, c.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("database init error: %w", err)
	}

	// 2. 初始化缓存 (Redis)
	redisCache, err := cache.NewRedisCache(c.Data.Redis, c.CircuitBreaker, logger, m)
	if err != nil {
		if sqlDB, err := db.RawDB().DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("redis init error: %w", err)
	}

	// 3. 初始化 Elasticsearch 客户端
	esClient, err := pkgsearch.NewClient(pkgsearch.Config{
		Addresses:     c.Data.Elasticsearch.Addresses,
		Username:      c.Data.Elasticsearch.Username,
		Password:      c.Data.Elasticsearch.Password,
		SlowThreshold: 500 * time.Millisecond,
		MaxRetries:    3,
		ServiceName:   BootstrapName,
		BreakerConfig: c.CircuitBreaker,
	}, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("elasticsearch init error: %w", err)
	}

	// 4. 初始化治理组件 (限流器、幂等管理器)
	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), c.RateLimit.Rate, time.Second)
	idemManager := idempotency.NewRedisManager(redisCache.GetClient(), IdempotencyPrefix)

	// 5. 初始化下游微服务客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, m, c.CircuitBreaker, clients)
	if err != nil {
		redisCache.Close()
		if sqlDB, err := db.RawDB().DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("grpc clients init error: %w", err)
	}

	// 6. DDD 分层装配
	bootLog.Info("assembling services with full dependency injection...")

	// 6.1 Infrastructure (Persistence)
	searchRepo := persistence.NewSearchRepository(db.RawDB())

	// 6.2 Application (Service)
	query := application.NewSearchQuery(searchRepo)
	manager := application.NewSearchManager(searchRepo, esClient, logger.Logger)
	searchService := application.NewSearch(manager, query, logger.Logger)

	// 7. 启动 Kafka 消费者进行可靠索引同步
	consumer := kafka.NewConsumer(c.MessageQueue.Kafka, logger, m)
	consumer.Start(context.Background(), 5, func(ctx context.Context, msg kafkago.Message) error {
		if msg.Topic != "product.index.sync" {
			return nil
		}
		var event map[string]any
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return err
		}

		productID := fmt.Sprintf("%v", event["product_id"])
		action := event["action"].(string)
		// --- 幂等保护：防止索引重复更新 ---
		idemKey := fmt.Sprintf("search:sync:%s:%s", productID, action)
		isFirst, _, err := idemManager.TryStart(ctx, idemKey, 1*time.Hour) // 索引同步时效性强，保留 1 小时即可
		if err != nil || !isFirst {
			return err
		}

		if err := manager.SyncProductIndex(ctx, event); err != nil {
			_ = idemManager.Delete(ctx, idemKey)
			return err
		}

		_ = idemManager.Finish(ctx, idemKey, &idempotency.Response{Body: "SYNCED"}, 1*time.Hour)
		return nil
	})

	// 6.3 Interface (HTTP Handlers)
	handler := searchhttp.NewHandler(searchService, logger.Logger)

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
		if sqlDB, err := db.RawDB().DB(); err == nil && sqlDB != nil {
			if err := sqlDB.Close(); err != nil {
				bootLog.Error("failed to close sql database", "error", err)
			}
		}
	}

	// 返回应用上下文与清理函数
	return &AppContext{
		Config:      c,
		Search:      searchService,
		Clients:     clients,
		Handler:     handler,
		Metrics:     m,
		Limiter:     rateLimiter,
		Idempotency: idemManager,
		Consumer:    consumer,
	}, cleanup, nil
}

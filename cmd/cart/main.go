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

	pb "github.com/wyfcoding/ecommerce/goapi/cart/v1"
	"github.com/wyfcoding/ecommerce/internal/cart/application"
	"github.com/wyfcoding/ecommerce/internal/cart/infrastructure/persistence"
	cartgrpc "github.com/wyfcoding/ecommerce/internal/cart/interfaces/grpc"
	carthttp "github.com/wyfcoding/ecommerce/internal/cart/interfaces/http"
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
)

// BootstrapName 服务唯一标识
const BootstrapName = "cart"

// IdempotencyPrefix 幂等性 Redis 键前缀
const IdempotencyPrefix = "cart:idem"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

// AppContext 应用上下文
type AppContext struct {
	Config      *Config
	Cart        *application.CartService
	Clients     *ServiceClients
	Handler     *carthttp.Handler
	Metrics     *metrics.Metrics
	Limiter     limiter.Limiter
	Idempotency idempotency.Manager
	Consumer    *kafka.Consumer
}

// ServiceClients 下游微服务客户端集合
type ServiceClients struct {
}

func main() {
	if err := app.NewBuilder(BootstrapName).
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

func registerGRPC(s *grpc.Server, svc any) {
	ctx := svc.(*AppContext)
	pb.RegisterCartServiceServer(s, cartgrpc.NewServer(ctx.Cart))
}

func registerGin(e *gin.Engine, svc any) {
	ctx := svc.(*AppContext)
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

func initService(cfg any, m *metrics.Metrics) (any, func(), error) {
	c := cfg.(*Config)
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default()

	// 打印脱敏配置
	configpkg.PrintWithMask(c)

	// 1. 初始化数据库 (MySQL)
	db, err := databases.NewDB(c.Data.Database, c.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, err
	}

	// 2. 初始化缓存
	redisCache, err := cache.NewRedisCache(c.Data.Redis, c.CircuitBreaker, logger, m)
	if err != nil {
		if sqlDB, err := db.RawDB().DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("redis init error: %w", err)
	}

	// 3. 初始化治理组件
	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), c.RateLimit.Rate, time.Second)
	idemManager := idempotency.NewRedisManager(redisCache.GetClient(), IdempotencyPrefix)

	// 4. 初始化下游微服务客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, m, c.CircuitBreaker, clients)
	if err != nil {
		redisCache.Close()
		if sqlDB, err := db.RawDB().DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("grpc clients init error: %w", err)
	}

	// 5. DDD 分层装配
	bootLog.Info("assembling services with full dependency injection...")

	// 5.1 Infrastructure (Persistence)
	cartRepo := persistence.NewCartRepository(db.RawDB())
	cartQuery := application.NewCartQuery(cartRepo, logger.Logger)
	cartManager := application.NewCartManager(cartRepo, logger.Logger, cartQuery)
	cartService := application.NewCartService(cartManager, cartQuery)

	// 5. [关键优化]：启动订单确认事件消费者，自动清空购物车
	consumer := kafka.NewConsumer(c.MessageQueue.Kafka, logger, m)
	consumer.Start(context.Background(), 5, func(ctx context.Context, msg kafkago.Message) error {
		if msg.Topic != "order.confirmed" {
			return nil
		}
		var event struct {
			OrderID uint64 `json:"order_id"`
			UserID  uint64 `json:"user_id"`
			Items   []struct {
				SkuID uint64 `json:"sku_id"`
			} `json:"items"`
		}
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return err
		}

		// 幂等处理：防止同一个订单重复清空购物车
		idemKey := fmt.Sprintf("cart:clear:order:%d", event.OrderID)
		isFirst, _, err := idemManager.TryStart(ctx, idemKey, 24*time.Hour)
		if err != nil || !isFirst {
			return err
		}

		skuIDs := make([]uint64, len(event.Items))
		for i, it := range event.Items {
			skuIDs[i] = it.SkuID
		}

		if err := cartManager.RemoveItems(ctx, event.UserID, skuIDs); err != nil {
			_ = idemManager.Delete(ctx, idemKey)
			return err
		}

		_ = idemManager.Finish(ctx, idemKey, &idempotency.Response{Body: "OK"}, 24*time.Hour)
		return nil
	})

	handler := carthttp.NewHandler(cartService, logger.Logger)

	cleanup := func() {
		bootLog.Info("shutting down, releasing resources...")
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
		Config: c, Cart: cartService, Handler: handler, Metrics: m,
		Limiter: rateLimiter, Idempotency: idemManager, Consumer: consumer,
	}, cleanup, nil
}

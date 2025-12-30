package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/goapi/order/v1"
	"github.com/wyfcoding/ecommerce/internal/order/application"
	"github.com/wyfcoding/ecommerce/internal/order/infrastructure/persistence"
	ordergrpc "github.com/wyfcoding/ecommerce/internal/order/interfaces/grpc"
	httpServer "github.com/wyfcoding/ecommerce/internal/order/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases/sharding"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/limiter"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
)

// BootstrapName 服务名称。
const BootstrapName = "order"

// Config 扩展配置。
type Config struct {
	configpkg.Config `mapstructure:",squash"`
	DTM              struct {
		Server string `mapstructure:"server"`
	} `mapstructure:"dtm"`
	Warehouse struct {
		GrpcAddr string `mapstructure:"grpc_addr"`
	} `mapstructure:"warehouse"`
}

// AppContext 应用上下文。
type AppContext struct {
	Config     *Config
	AppService *application.OrderService
	Clients    *ServiceClients
	Handler    *httpServer.Handler
	Metrics    *metrics.Metrics
	Limiter    limiter.Limiter
}

// ServiceClients 下游微服务。
type ServiceClients struct {
	Warehouse *grpc.ClientConn `service:"warehouse"`
}

func main() {
	if err := app.NewBuilder(BootstrapName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGinMiddleware(
			middleware.MetricsMiddleware(),
			middleware.CORS(),
		).
		Build().
		Run(); err != nil {
		slog.Error("service bootstrap failed", "error", err)
	}
}

func registerGRPC(s *grpc.Server, srv any) {
	ctx := srv.(*AppContext)
	pb.RegisterOrderServiceServer(s, ordergrpc.NewServer(ctx.AppService))
}

func registerGin(e *gin.Engine, srv any) {
	ctx := srv.(*AppContext)

	if ctx.Config.Server.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 1. 基础路由层 (跳过限流)
	sys := e.Group("/sys")
	{
		sys.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "UP",
				"service":   BootstrapName,
				"timestamp": time.Now().Unix(),
			})
		})
		sys.GET("/ready", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "READY"})
		})
	}

	if ctx.Config.Metrics.Enabled {
		e.GET(ctx.Config.Metrics.Path, gin.WrapH(ctx.Metrics.Handler()))
	}

	// 2. 业务限流
	e.Use(middleware.RateLimitWithLimiter(ctx.Limiter))

	// 3. 业务路由
	ctx.Handler.RegisterRoutes(e.Group("/api/v1"))

	slog.Info("HTTP service configured", "service", BootstrapName)
}

func initService(cfg any, m *metrics.Metrics) (any, func(), error) {
	c := cfg.(*Config)
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default()

	// 1. 基础设施：分片数据库管理器
	bootLog.Info("initializing database manager...")
	var (
		shardingManager *sharding.Manager
		err             error
	)

	if len(c.Data.Shards) > 0 {
		shardingManager, err = sharding.NewManager(c.Data.Shards, logger)
	} else {
		// 容错：如果未配置分片，使用单库配置构造 Manager
		shardingManager, err = sharding.NewManager([]configpkg.DatabaseConfig{c.Data.Database}, logger)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("database manager init failed: %w", err)
	}

	// 2. Redis
	redisCache, err := cache.NewRedisCache(c.Data.Redis, logger)
	if err != nil {
		if closeErr := shardingManager.Close(); closeErr != nil {
			bootLog.Error("failed to close sharding manager during error recovery", "error", closeErr)
		}
		return nil, nil, fmt.Errorf("redis connection failed: %w", err)
	}

	// 3. 限流器
	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), c.RateLimit.Rate, time.Second)

	// 4. ID 生成器 & 消息队列
	idGen, err := idgen.NewSnowflakeGenerator(c.Snowflake)
	if err != nil {
		if closeErr := redisCache.Close(); closeErr != nil {
			bootLog.Error("failed to close redis cache during error recovery", "error", closeErr)
		}
		if closeErr := shardingManager.Close(); closeErr != nil {
			bootLog.Error("failed to close sharding manager during error recovery", "error", closeErr)
		}
		return nil, nil, fmt.Errorf("idgen init failed: %w", err)
	}
	kafkaProducer := kafka.NewProducer(c.MessageQueue.Kafka, logger)

	// 5. 下游客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, clients)
	if err != nil {
		if closeErr := kafkaProducer.Close(); closeErr != nil {
			bootLog.Error("failed to close kafka producer during error recovery", "error", closeErr)
		}
		if closeErr := redisCache.Close(); closeErr != nil {
			bootLog.Error("failed to close redis cache during error recovery", "error", closeErr)
		}
		if closeErr := shardingManager.Close(); closeErr != nil {
			bootLog.Error("failed to close sharding manager during error recovery", "error", closeErr)
		}
		return nil, nil, fmt.Errorf("grpc clients init failed: %w", err)
	}

	// 6. DDD 组装
	bootLog.Info("assembling order application service...")
	repo := persistence.NewOrderRepository(shardingManager)
	orderManager := application.NewOrderManager(repo, idGen, kafkaProducer, logger.Logger, c.DTM.Server, c.Warehouse.GrpcAddr, m)
	orderQuery := application.NewOrderQuery(repo)

	service := application.NewOrderService(orderManager, orderQuery, logger.Logger)
	handler := httpServer.NewHandler(service, logger.Logger)

	cleanup := func() {
		bootLog.Info("performing graceful shutdown...")
		clientCleanup()
		if err := kafkaProducer.Close(); err != nil {
			bootLog.Error("failed to close kafka producer", "error", err)
		}
		if err := redisCache.Close(); err != nil {
			bootLog.Error("failed to close redis cache", "error", err)
		}
		if err := shardingManager.Close(); err != nil {
			bootLog.Error("failed to close sharding manager", "error", err)
		}
	}

	return &AppContext{
		Config:     c,
		AppService: service,
		Clients:    clients,
		Handler:    handler,
		Metrics:    m,
		Limiter:    rateLimiter,
	}, cleanup, nil
}

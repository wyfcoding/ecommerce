package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/goapi/payment/v1"
	settlementv1 "github.com/wyfcoding/ecommerce/goapi/settlement/v1"
	"github.com/wyfcoding/ecommerce/internal/payment/application"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/ecommerce/internal/payment/infrastructure/gateway"
	"github.com/wyfcoding/ecommerce/internal/payment/infrastructure/persistence"
	"github.com/wyfcoding/ecommerce/internal/payment/infrastructure/risk"
	grpcServer "github.com/wyfcoding/ecommerce/internal/payment/interfaces/grpc"
	paymenthttp "github.com/wyfcoding/ecommerce/internal/payment/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases/sharding"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/idempotency"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/limiter"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
)

// BootstrapName 服务唯一标识
const BootstrapName = "payment"

// IdempotencyPrefix 幂等性 Redis 键前缀
const IdempotencyPrefix = "payment:idem"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

// AppContext 应用上下文 (包含对外服务实例与依赖)
type AppContext struct {
	Config      *Config
	Payment     *application.PaymentService
	Clients     *ServiceClients
	Handler     *paymenthttp.Handler
	Metrics     *metrics.Metrics
	Limiter     limiter.Limiter
	Idempotency idempotency.Manager
}

// ServiceClients 下游微服务客户端集合
type ServiceClients struct {
	SettlementConn *grpc.ClientConn `service:"settlement"`
	OrderConn      *grpc.ClientConn `service:"order"`

	// 具体的客户端接口 (由 Conn 转化)
	Settlement settlementv1.SettlementServiceClient
}

func main() {
	// 构建并运行服务
	if err := app.NewBuilder(BootstrapName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGinMiddleware(
			middleware.MetricsMiddleware(),               // 指标采集
			middleware.CORS(),                            // 跨域处理
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
	pb.RegisterPaymentServiceServer(s, grpcServer.NewServer(ctx.Payment))
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
		// 支付接口通常需要严格鉴权
		api.Use(middleware.JWTAuth(ctx.Config.JWT.Secret))
		// 支付核心接口必须保证幂等
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
		shardingManager, err = sharding.NewManager(c.Data.Shards, logger)
	} else {
		shardingManager, err = sharding.NewManager([]configpkg.DatabaseConfig{c.Data.Database}, logger)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("sharding database init error: %w", err)
	}

	// 2. 初始化缓存 (Redis)
	redisCache, err := cache.NewRedisCache(c.Data.Redis, logger)
	if err != nil {
		shardingManager.Close()
		return nil, nil, fmt.Errorf("redis init error: %w", err)
	}

	// 3. 初始化治理组件 (限流器、幂等管理器、ID 生成器)
	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), c.RateLimit.Rate, time.Second)
	idemManager := idempotency.NewRedisManager(redisCache.GetClient(), IdempotencyPrefix)
	idGenerator, _ := idgen.NewGenerator(c.Snowflake)

	// 4. 初始化下游微服务客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, clients)
	if err != nil {
		redisCache.Close()
		shardingManager.Close()
		return nil, nil, fmt.Errorf("grpc clients init error: %w", err)
	}
	// 显式转换 gRPC 客户端
	if clients.SettlementConn != nil {
		clients.Settlement = settlementv1.NewSettlementServiceClient(clients.SettlementConn)
	}

	// 5. DDD 分层装配
	bootLog.Info("assembling services with full dependency injection...")

	// 5.1 Infrastructure (Persistence, Gateways, Risk)
	paymentRepo := persistence.NewPaymentRepository(shardingManager)
	channelRepo := persistence.NewChannelRepository(shardingManager)
	refundRepo := persistence.NewRefundRepository(shardingManager)

	riskSvc := risk.NewRiskService()

	gateways := map[domain.GatewayType]domain.PaymentGateway{
		domain.GatewayTypeAlipay: gateway.NewAlipayGateway(),
		domain.GatewayTypeStripe: gateway.NewStripeGateway(),
		domain.GatewayTypeWechat: gateway.NewWechatGateway(),
	}

	// 5.2 Application (Components)
	processor := application.NewPaymentProcessor(
		paymentRepo,
		channelRepo,
		riskSvc,
		idGenerator,
		gateways,
		logger.Logger,
	)
	callbackHandler := application.NewCallbackHandler(paymentRepo, gateways, logger.Logger)
	refundService := application.NewRefundService(paymentRepo, refundRepo, idGenerator, gateways, logger.Logger)
	paymentQuery := application.NewPaymentQuery(paymentRepo)

	paymentService := application.NewPaymentService(
		processor,
		callbackHandler,
		refundService,
		paymentQuery,
		clients.Settlement,
		logger.Logger,
	)

	// 5.3 Interface (HTTP Handlers)
	handler := paymenthttp.NewHandler(paymentService, logger.Logger)

	// 定义资源清理函数
	cleanup := func() {
		bootLog.Info("shutting down, releasing resources...")
		clientCleanup()
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
		Payment:     paymentService,
		Clients:     clients,
		Handler:     handler,
		Metrics:     m,
		Limiter:     rateLimiter,
		Idempotency: idemManager,
	}, cleanup, nil
}

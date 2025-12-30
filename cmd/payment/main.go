package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	v1 "github.com/wyfcoding/ecommerce/goapi/payment/v1"
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
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/limiter"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
)

// BootstrapName 服务标识。
const BootstrapName = "payment"

// Config 扩展配置。
type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

// AppContext 应用上下文。
type AppContext struct {
	Config     *Config
	AppService *application.PaymentService
	Clients    *ServiceClients
	Metrics    *metrics.Metrics
	Limiter    limiter.Limiter
}

// ServiceClients 下游微服务。
type ServiceClients struct {
	Settlement *grpc.ClientConn `service:"settlement"`
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

func registerGRPC(s *grpc.Server, svc any) {
	ctx := svc.(*AppContext)
	v1.RegisterPaymentServiceServer(s, grpcServer.NewServer(ctx.AppService))
}

func registerGin(e *gin.Engine, svc any) {
	ctx := svc.(*AppContext)

	if ctx.Config.Server.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 1. 系统路由 (不限流)
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

	// 2. 支付链路保护：限流
	e.Use(middleware.RateLimitWithLimiter(ctx.Limiter))

	// 3. 业务路由
	handler := paymenthttp.NewHandler(ctx.AppService, slog.Default())
	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Info("HTTP service configured", "service", BootstrapName)
}

func initService(cfg any, m *metrics.Metrics) (any, func(), error) {
	c := cfg.(*Config)
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default()

	// 1. 基础设施：分片数据库
	shardingManager, err := sharding.NewManager(c.Data.Shards, logger)
	if err != nil {
		if len(c.Data.Shards) == 0 {
			shardingManager, err = sharding.NewManager([]configpkg.DatabaseConfig{c.Data.Database}, logger)
			if err != nil {
				return nil, nil, fmt.Errorf("sharding manager init error: %w", err)
			}
		} else {
			return nil, nil, fmt.Errorf("sharding manager init error: %w", err)
		}
	}

	// 2. Redis & 限流器
	redisCache, err := cache.NewRedisCache(c.Data.Redis, logger)
	if err != nil {
		shardingManager.Close()
		return nil, nil, fmt.Errorf("redis init error: %w", err)
	}
	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), c.RateLimit.Rate, time.Second)

	// 3. ID 生成器
	idGenerator, err := idgen.NewSnowflakeGenerator(c.Snowflake)
	if err != nil {
		redisCache.Close()
		shardingManager.Close()
		return nil, nil, fmt.Errorf("idgen init error: %w", err)
	}

	// 4. 下游客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, clients)
	if err != nil {
		redisCache.Close()
		shardingManager.Close()
		return nil, nil, fmt.Errorf("clients init error: %w", err)
	}

	// 5. 业务装配
	bootLog.Info("assembling payment application service...")
	paymentRepo := persistence.NewPaymentRepository(shardingManager)
	refundRepo := persistence.NewRefundRepository(shardingManager)
	channelRepo := persistence.NewChannelRepository(shardingManager)
	riskService := risk.NewRiskService()

	gateways := map[domain.GatewayType]domain.PaymentGateway{
		domain.GatewayTypeAlipay: gateway.NewAlipayGateway(),
		domain.GatewayTypeWechat: gateway.NewWechatGateway(),
		domain.GatewayTypeStripe: gateway.NewStripeGateway(),
		domain.GatewayTypeMock:   gateway.NewAlipayGateway(),
	}

	var settlementCli settlementv1.SettlementServiceClient
	if clients.Settlement != nil {
		settlementCli = settlementv1.NewSettlementServiceClient(clients.Settlement)
	}

	processor := application.NewPaymentProcessor(paymentRepo, channelRepo, riskService, idGenerator, gateways, logger.Logger)
	callbackHandler := application.NewCallbackHandler(paymentRepo, gateways, logger.Logger)
	refundService := application.NewRefundService(paymentRepo, refundRepo, idGenerator, gateways, logger.Logger)
	paymentQuery := application.NewPaymentQuery(paymentRepo)

	appService := application.NewPaymentService(processor, callbackHandler, refundService, paymentQuery, settlementCli, logger.Logger)

	cleanup := func() {
		bootLog.Info("performing graceful shutdown...")
		clientCleanup()
		redisCache.Close()
		shardingManager.Close()
	}

	return &AppContext{
		Config:     c,
		AppService: appService,
		Clients:    clients,
		Metrics:    m,
		Limiter:    rateLimiter,
	}, cleanup, nil
}

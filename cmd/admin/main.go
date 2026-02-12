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

	pb "github.com/wyfcoding/ecommerce/go-api/admin/v1"
	"github.com/wyfcoding/ecommerce/internal/admin/application"
	"github.com/wyfcoding/ecommerce/internal/admin/domain"
	admes "github.com/wyfcoding/ecommerce/internal/admin/infrastructure/persistence/elasticsearch"
	admysql "github.com/wyfcoding/ecommerce/internal/admin/infrastructure/persistence/mysql"
	adredis "github.com/wyfcoding/ecommerce/internal/admin/infrastructure/persistence/redis"
	adminconsumer "github.com/wyfcoding/ecommerce/internal/admin/interfaces/consumer"
	admingrpc "github.com/wyfcoding/ecommerce/internal/admin/interfaces/grpc"
	adminhttp "github.com/wyfcoding/ecommerce/internal/admin/interfaces/http"
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
	"github.com/wyfcoding/pkg/storage"
)

// BootstrapName 服务唯一标识
const BootstrapName = "admin"

// IdempotencyPrefix 幂等性 Redis 键前缀
const IdempotencyPrefix = "admin:idem"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
	Search           struct {
		AuditLogIndex string `mapstructure:"audit_log_index" toml:"audit_log_index"`
	} `mapstructure:"search" toml:"search"`
}

// AppContext 应用上下文 (包含对外服务实例与依赖)
type AppContext struct {
	Config      *Config
	Command     *application.AdminCommandService
	Query       *application.AdminQueryService
	Clients     *ServiceClients
	Handler     *adminhttp.AdminHandler
	Metrics     *metrics.Metrics
	Limiter     limiter.Limiter
	Idempotency idempotency.Manager
}

// ServiceClients 下游微服务客户端集合
type ServiceClients struct {
	User    *grpc.ClientConn `service:"user"`
	Order   *grpc.ClientConn `service:"order"`
	Payment *grpc.ClientConn `service:"payment"`
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
			middleware.TimeoutMiddleware(30*time.Second), // 全局超时 (注: 此处无法读取配置，使用默认值)
		).
		Build().
		Run(); err != nil {
		slog.Error("service bootstrap failed", "error", err)
	}
}

// registerGRPC 注册 gRPC 服务
func registerGRPC(s *grpc.Server, ctx *AppContext) {
	pb.RegisterAdminServiceServer(s, admingrpc.NewServer(ctx.Command, ctx.Query))
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
		// 公开与鉴权接口统一由 Handler 注册
		ctx.Handler.RegisterRoutes(api, ctx.Config.JWT.Secret)
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

	// 2.1 初始化 Elasticsearch 客户端 (审计日志搜索)
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
	outboxProcessor := outbox.NewProcessor(outboxMgr, func(ctx context.Context, topic, key string, payload []byte) error {
		return producer.PublishToTopic(ctx, topic, []byte(key), payload)
	}, 100, 5*time.Second)
	outboxProcessor.Start()

	// 4. 初始化存储基础设施 (MinIO)
	// 注意: 作为通用能力注入到 Service，而非作为 Repository
	store, err := storage.NewMinIOClient(
		c.Minio.Endpoint,
		c.Minio.AccessKeyID,
		c.Minio.SecretAccessKey,
		c.Minio.BucketName,
		c.Minio.UseSSL,
	)
	if err != nil {
		bootLog.Warn("storage init failed, continuing without storage", "error", err)
	}

	// 5. 初始化下游微服务客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, m, c.CircuitBreaker, clients)
	if err != nil {
		outboxProcessor.Stop()
		producer.Close()
		redisCache.Close()
		if sqlDB, err := db.RawDB().DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("grpc clients init error: %w", err)
	}

	// 6. DDD 分层装配 (Infrastructure -> Domain -> Application -> Interface)
	bootLog.Info("assembling services with full dependency injection...")

	// 6.1 Infrastructure (Persistence)
	adminRepo := admysql.NewAdminRepository(db.RawDB())
	roleRepo := admysql.NewRoleRepository(db.RawDB())
	auditRepo := admysql.NewAuditRepository(db.RawDB())
	approvalRepo := admysql.NewApprovalRepository(db.RawDB())
	settingRepo := admysql.NewSettingRepository(db.RawDB())

	adminReadRepo := adredis.NewAdminUserReadRepository(redisCache.GetClient(), c.Cache.DefaultExpiration)
	settingReadRepo := adredis.NewSettingReadRepository(redisCache.GetClient(), c.Cache.DefaultExpiration)
	auditSearchRepo := admes.NewAuditLogSearchRepository(esClient, c.Search.AuditLogIndex)

	// 6.2 Application (Service)
	// 注入外部依赖 (Parameter Object Pattern)
	opsDeps := application.SystemOpsDependencies{
		OrderClient:   clients.Order,
		UserClient:    clients.User,
		PaymentClient: clients.Payment,
		Storage:       store,
	}

	commandSvc := application.NewAdminCommandService(
		adminRepo,
		roleRepo,
		auditRepo,
		settingRepo,
		approvalRepo,
		opsDeps,
		outbox.NewPublisher(outboxMgr),
		logger.Logger,
	)
	querySvc := application.NewAdminQueryService(
		adminRepo,
		roleRepo,
		auditRepo,
		settingRepo,
		approvalRepo,
		adminReadRepo,
		settingReadRepo,
		auditSearchRepo,
	)

	// 6.3 Projection Consumers (Admin Events -> Read Model)
	projectionService := application.NewAdminProjectionService(adminRepo, adminReadRepo, settingRepo, settingReadRepo, auditRepo, auditSearchRepo, logger.Logger)
	projectionHandler := adminconsumer.NewAdminProjectionHandler(projectionService, logger.Logger)
	projectionTopics := []string{
		domain.AdminUserCreatedEventType,
		domain.AdminUserUpdatedEventType,
		domain.AdminUserDisabledEventType,
		domain.RoleAssignedEventType,
		domain.SystemSettingUpdatedEventType,
		domain.AuditLogCreatedEventType,
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

	// 6.4 Interface (HTTP Handlers)
	handler := adminhttp.NewAdminHandler(commandSvc, querySvc, logger.Logger)

	// 定义资源清理函数
	cleanup := func() {
		bootLog.Info("shutting down, releasing resources...")
		for _, c := range projectionConsumers {
			if c != nil {
				c.Close()
			}
		}
		outboxProcessor.Stop()
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
		Command:     commandSvc,
		Query:       querySvc,
		Clients:     clients,
		Handler:     handler,
		Metrics:     m,
		Limiter:     rateLimiter,
		Idempotency: idemManager,
	}, cleanup, nil
}

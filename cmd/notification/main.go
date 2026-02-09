package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/goapi/notification/v1"
	"github.com/wyfcoding/ecommerce/internal/notification/application"
	"github.com/wyfcoding/ecommerce/internal/notification/domain"
	notificationsearch "github.com/wyfcoding/ecommerce/internal/notification/infrastructure/persistence/elasticsearch"
	notificationmysql "github.com/wyfcoding/ecommerce/internal/notification/infrastructure/persistence/mysql"
	notificationredis "github.com/wyfcoding/ecommerce/internal/notification/infrastructure/persistence/redis"
	notificationconsumer "github.com/wyfcoding/ecommerce/internal/notification/interfaces/consumer"
	notificationgrpc "github.com/wyfcoding/ecommerce/internal/notification/interfaces/grpc"
	notificationhttp "github.com/wyfcoding/ecommerce/internal/notification/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/idempotency"
	"github.com/wyfcoding/pkg/limiter"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/messagequeue/outbox"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
	"github.com/wyfcoding/pkg/notification"
	"github.com/wyfcoding/pkg/response"
	"github.com/wyfcoding/pkg/search"
	"github.com/wyfcoding/pkg/server"
)

// BootstrapName 服务唯一标识
const BootstrapName = "notification"

// IdempotencyPrefix 幂等性 Redis 键前缀
const IdempotencyPrefix = "notification:idem"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
	Search           struct {
		NotificationIndex string `mapstructure:"notification_index" toml:"notification_index"`
		TemplateIndex     string `mapstructure:"template_index" toml:"template_index"`
	} `mapstructure:"search" toml:"search"`

	Email   notification.EmailConfig   `mapstructure:"email" toml:"email"`
	SMS     notification.SMSConfig     `mapstructure:"sms" toml:"sms"`
	Webhook notification.WebhookConfig `mapstructure:"webhook" toml:"webhook"`
}

// AppContext 应用上下文 (包含对外服务实例与依赖)
type AppContext struct {
	Config       *Config
	Cmd          *application.NotificationCommandService
	Query        *application.NotificationQueryService
	Clients      *ServiceClients
	Handler      *notificationhttp.Handler
	WsHandler    *notificationhttp.WebsocketHandler
	WebsocketMgr *server.WSManager
	Metrics      *metrics.Metrics
	Limiter      limiter.Limiter
	Idempotency  idempotency.Manager
}

// ServiceClients 下游微服务客户端集合
type ServiceClients struct {
	// 目前 Notification 服务无下游强依赖
}

func main() {
	// 构建并运行服务
	if err := app.NewBuilder[*Config, *AppContext](BootstrapName).
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

// registerGRPC 注册 gRPC 服务
func registerGRPC(s *grpc.Server, ctx *AppContext) {
	pb.RegisterNotificationServiceServer(s, notificationgrpc.NewServer(ctx.Cmd, ctx.Query))
}

// registerGin 注册 HTTP 路由
func registerGin(e *gin.Engine, ctx *AppContext) {
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
		ctx.Handler.RegisterRoutes(api)
		ctx.WsHandler.RegisterRoutes(api)
	}
}

// initService 初始化服务依赖 (数据库、缓存、客户端、领域层)
func initService(cfg *Config, m *metrics.Metrics) (*AppContext, func(), error) {
	c := cfg
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default()

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

	// 4. 初始化下游微服务客户端
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

	// 5. DDD 分层装配
	bootLog.Info("assembling services with full dependency injection...")

	// 5.1 Infrastructure (Persistence & Senders)
	notificationRepo := notificationmysql.NewNotificationRepository(db.RawDB())
	notificationReadRepo := notificationredis.NewNotificationReadRepository(redisCache.GetClient(), c.Cache.DefaultExpiration)
	templateReadRepo := notificationredis.NewNotificationTemplateReadRepository(redisCache.GetClient(), c.Cache.DefaultExpiration)
	notificationSearchRepo := notificationsearch.NewNotificationSearchRepository(esClient, c.Search.NotificationIndex)
	templateSearchRepo := notificationsearch.NewNotificationTemplateSearchRepository(esClient, c.Search.TemplateIndex)

	// 初始化真实 Kafka 发送器
	emailSender := kafka.NewNotificationSender(producer, "notification.email")
	smsSender := kafka.NewNotificationSender(producer, "notification.sms")
	webhookSender := application.NewWebhookSender()

	// 5.2 Application (Service)
	// 初始化 WebsocketManager (复用 pkg/server)
	websocketMgr := server.NewWSManager(logger.Logger)
	go websocketMgr.Run(context.Background())

	commandSvc := application.NewNotificationCommandService(notificationRepo, templateReadRepo, outbox.NewPublisher(outboxMgr), emailSender, smsSender, webhookSender, websocketMgr, logger.Logger)
	querySvc := application.NewNotificationQueryService(notificationRepo, notificationReadRepo, templateReadRepo, notificationSearchRepo, templateSearchRepo, logger.Logger)

	// 5.3 Projection Consumers (Notification Events -> Read Model)
	projectionService := application.NewNotificationProjectionService(notificationRepo, notificationReadRepo, templateReadRepo, notificationSearchRepo, templateSearchRepo, logger.Logger)
	projectionHandler := notificationconsumer.NewNotificationProjectionHandler(projectionService, logger.Logger)
	projectionTopics := []string{
		domain.NotificationCreatedEventType,
		domain.NotificationReadEventType,
		domain.NotificationDeletedEventType,
		domain.NotificationTemplateCreatedEventType,
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

	// 5.5 Notification Worker Consumers (Actual Sending)
	// 初始化真实发送器 SDK
	pkgEmailSender := notification.NewEmailSender(&c.Email, logger.Logger)
	pkgSMSSender := notification.NewSMSSender(&c.SMS, logger.Logger)
	pkgWebhookSender := notification.NewWebhookSender(&c.Webhook, logger.Logger)

	// 初始化处理器
	emailHandler := notificationconsumer.NewNotificationSenderHandler(pkgEmailSender, logger.Logger)
	smsHandler := notificationconsumer.NewNotificationSenderHandler(pkgSMSSender, logger.Logger)
	webhookHandler := notificationconsumer.NewNotificationSenderHandler(pkgWebhookSender, logger.Logger)

	// 启动消费者
	workerTopics := map[string]*notificationconsumer.NotificationSenderHandler{
		"notification.email":   emailHandler,
		"notification.sms":     smsHandler,
		"notification.webhook": webhookHandler,
	}

	for topic, handler := range workerTopics {
		consumerCfg := c.MessageQueue.Kafka
		consumerCfg.Topic = topic
		consumerCfg.GroupID = BootstrapName + "-worker-group" // 使用独立消费组
		consumer := kafka.NewConsumer(&consumerCfg, logger, m)
		consumer.Start(context.Background(), 5, handler.Handle)
		projectionConsumers = append(projectionConsumers, consumer) // 加入清理列表
	}

	// 5.4 Interface (HTTP Handlers)
	handler := notificationhttp.NewHandler(commandSvc, querySvc, logger.Logger)
	wsHandler := notificationhttp.NewWebsocketHandler(websocketMgr, c.JWT.Secret, logger.Logger)

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

	return &AppContext{
		Config:       c,
		Cmd:          commandSvc,
		Query:        querySvc,
		Clients:      clients,
		Handler:      handler,
		WsHandler:    wsHandler,
		WebsocketMgr: websocketMgr,
		Metrics:      m,
		Limiter:      rateLimiter,
		Idempotency:  idemManager,
	}, cleanup, nil
}

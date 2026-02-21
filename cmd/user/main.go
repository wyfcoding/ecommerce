// 变更说明：重构用户服务装配，统一 DDD/CQRS + Outbox + Redis/ES 读模型。
package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/user/v1"
	"github.com/wyfcoding/ecommerce/internal/user/application"
	"github.com/wyfcoding/ecommerce/internal/user/domain"
	usersearch "github.com/wyfcoding/ecommerce/internal/user/infrastructure/persistence/elasticsearch"
	usermysql "github.com/wyfcoding/ecommerce/internal/user/infrastructure/persistence/mysql"
	userredis "github.com/wyfcoding/ecommerce/internal/user/infrastructure/persistence/redis"
	usergrpc "github.com/wyfcoding/ecommerce/internal/user/interfaces/grpc"
	userhttp "github.com/wyfcoding/ecommerce/internal/user/interfaces/http"
	"github.com/wyfcoding/pkg/algos/infra"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/limiter"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/messagequeue/outbox"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
	"github.com/wyfcoding/pkg/search"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

// BootstrapName 服务唯一标识
const BootstrapName = "user"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
	Search           struct {
		UserIndex string `mapstructure:"user_index" toml:"user_index"`
	} `mapstructure:"search" toml:"search"`
}

// AppContext 应用上下文
type AppContext struct {
	Config  *Config
	Cmd     *application.UserCommandService
	Query   *application.UserQueryService
	Handler *userhttp.UserHandler
	Metrics *metrics.Metrics
	Limiter limiter.Limiter
}

func main() {
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

func registerGRPC(s *grpc.Server, ctx *AppContext) {
	pb.RegisterUserServiceServer(s, usergrpc.NewGrpcHandler(ctx.Cmd, ctx.Query))
}

func registerGin(e *gin.Engine, ctx *AppContext) {
	if ctx.Config.Server.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	if ctx.Limiter != nil {
		e.Use(middleware.RateLimitWithLimiter(ctx.Limiter))
	}
	ctx.Handler.RegisterHandlers(e)
}

func initService(cfg *Config, m *metrics.Metrics) (*AppContext, func(), error) {
	c := cfg
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default()

	configpkg.PrintWithMask(c)

	// 1. DB
	db, err := database.NewDB(c.Data.Database, c.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("database init error: %w", err)
	}
	if c.Server.Environment == "dev" {
		if err := db.RawDB().AutoMigrate(&usermysql.UserModel{}, &usermysql.AddressModel{}, &outbox.Message{}); err != nil {
			bootLog.Error("failed to migrate database", "error", err)
		}
	}

	// 2. Redis
	redisCache, err := cache.NewRedisCache(&c.Data.Redis, c.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("redis init error: %w", err)
	}

	// 3. Kafka & Outbox
	producer := kafka.NewProducer(&c.MessageQueue.Kafka, logger, m)
	outboxMgr := outbox.NewManager(db.RawDB(), logger.Logger)
	outboxProcessor := outbox.NewProcessor(outboxMgr, func(ctx context.Context, topic, key string, payload []byte) error {
		return producer.PublishToTopic(ctx, topic, []byte(key), payload)
	}, 100, 5*time.Second)
	outboxProcessor.Start()

	// 4. ES (optional)
	var searchRepo domain.UserSearchRepository
	esClient, err := search.NewClient(&search.Config{
		ServiceName:         BootstrapName,
		ElasticsearchConfig: c.Data.Elasticsearch,
		BreakerConfig:       c.CircuitBreaker,
		SlowThreshold:       800 * time.Millisecond,
		MaxRetries:          3,
	}, logger, m)
	if err != nil {
		bootLog.Warn("elasticsearch init skipped", "error", err)
	} else {
		searchRepo = usersearch.NewUserSearchRepository(esClient, c.Search.UserIndex)
	}

	// 5. Repositories
	userRepo := usermysql.NewUserRepository(db.RawDB())
	addressRepo := usermysql.NewAddressRepository(db.RawDB())
	userReadRepo := userredis.NewUserReadRepository(redisCache.GetClient(), time.Hour)
	addressReadRepo := userredis.NewAddressReadRepository(redisCache.GetClient(), 30*time.Minute)

	// 6. Application Services
	antiBot := infra.NewAntiBotDetector()
	publisher := outbox.NewPublisher(outboxMgr)
	cmdSvc := application.NewUserCommandService(
		userRepo,
		addressRepo,
		userReadRepo,
		addressReadRepo,
		searchRepo,
		publisher,
		c.MessageQueue.Kafka.Topic,
		c.JWT.Secret,
		c.JWT.Issuer,
		c.JWT.ExpireDuration,
		antiBot,
		logger.Logger,
	)
	querySvc := application.NewUserQueryService(
		userRepo,
		addressRepo,
		userReadRepo,
		addressReadRepo,
		searchRepo,
		antiBot,
		logger.Logger,
	)

	handler := userhttp.NewUserHandler(cmdSvc, querySvc)

	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), c.RateLimit.Rate, c.RateLimit.Burst)

	cleanup := func() {
		bootLog.Info("shutting down, releasing resources...")
		outboxProcessor.Stop()
		if producer != nil {
			producer.Close()
		}
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
		Config:  c,
		Cmd:     cmdSvc,
		Query:   querySvc,
		Handler: handler,
		Metrics: m,
		Limiter: rateLimiter,
	}, cleanup, nil
}

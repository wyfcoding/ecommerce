// 变更说明：重构 KYC 服务，使用 DDD/CQRS + Outbox + Redis 读模型架构
package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/kyc/v1"
	"github.com/wyfcoding/ecommerce/internal/kyc/application"
	"github.com/wyfcoding/ecommerce/internal/kyc/domain"
	kycmysql "github.com/wyfcoding/ecommerce/internal/kyc/infrastructure/persistence/mysql"
	kycredis "github.com/wyfcoding/ecommerce/internal/kyc/infrastructure/persistence/redis"
	kycgrpc "github.com/wyfcoding/ecommerce/internal/kyc/interfaces/grpc"
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

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

// BootstrapName 服务唯一标识
const BootstrapName = "kyc"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

// AppContext 应用上下文
type AppContext struct {
	Config  *Config
	Cmd     *application.KYCCommandService
	Query   *application.KYCQueryService
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
	pb.RegisterKYCServiceServer(s, kycgrpc.NewKYCHandler(ctx.Cmd, ctx.Query))
}

func registerGin(e *gin.Engine, ctx *AppContext) {
	if ctx.Config.Server.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	if ctx.Limiter != nil {
		e.Use(middleware.RateLimitWithLimiter(ctx.Limiter))
	}
	// KYC 服务主要通过 gRPC 提供服务，HTTP 仅提供健康检查
	e.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
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
		if err := db.RawDB().AutoMigrate(
			&kycmysql.KYCApplicationModel{},
			&kycmysql.DocumentModel{},
			&kycmysql.FaceVerificationModel{},
			&kycmysql.AuditRecordModel{},
			&kycmysql.MerchantKYCModel{},
			&outbox.Message{},
		); err != nil {
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

	// 4. Repositories
	kycRepo := kycmysql.NewKYCRepository(db.RawDB())
	docRepo := kycmysql.NewDocumentRepository(db.RawDB())
	faceRepo := kycmysql.NewFaceVerificationRepository(db.RawDB())
	auditRepo := kycmysql.NewAuditRecordRepository(db.RawDB())
	merchantRepo := kycmysql.NewMerchantKYCRepository(db.RawDB())
	readRepo := kycredis.NewKYCReadRepository(redisCache.GetClient(), time.Hour)

	// 5. Domain Service (OCR、人脸识别等服务可以后续集成)
	domainService := domain.NewKYCDomainService(nil, nil, nil, nil, nil)

	// 6. Application Services
	publisher := outbox.NewPublisher(outboxMgr)
	cmdSvc := application.NewKYCCommandService(
		kycRepo,
		docRepo,
		faceRepo,
		auditRepo,
		merchantRepo,
		readRepo,
		domainService,
		publisher,
		c.MessageQueue.Kafka.Topic,
		logger.Logger,
	)
	querySvc := application.NewKYCQueryService(
		kycRepo,
		docRepo,
		faceRepo,
		auditRepo,
		merchantRepo,
		readRepo,
		logger.Logger,
	)

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
		Metrics: m,
		Limiter: rateLimiter,
	}, cleanup, nil
}

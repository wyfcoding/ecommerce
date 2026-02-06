package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/goapi/product/v1"
	"github.com/wyfcoding/ecommerce/internal/product/application"
	"github.com/wyfcoding/ecommerce/internal/product/domain"
	productsearch "github.com/wyfcoding/ecommerce/internal/product/infrastructure/persistence/elasticsearch"
	productmysql "github.com/wyfcoding/ecommerce/internal/product/infrastructure/persistence/mysql"
	productredis "github.com/wyfcoding/ecommerce/internal/product/infrastructure/persistence/redis"
	productconsumer "github.com/wyfcoding/ecommerce/internal/product/interfaces/consumer"
	productgrpc "github.com/wyfcoding/ecommerce/internal/product/interfaces/grpc"
	producthttp "github.com/wyfcoding/ecommerce/internal/product/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/messagequeue/outbox"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
	"github.com/wyfcoding/pkg/search"
)

// BootstrapName 服务唯一标识
const BootstrapName = "product"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
	Search           struct {
		ProductIndex string `mapstructure:"product_index" toml:"product_index"`
	} `mapstructure:"search" toml:"search"`
}

// AppContext 应用上下文
type AppContext struct {
	Config  *Config
	Cmd     *application.ProductCommandService
	Query   *application.ProductQueryService
	Handler *producthttp.Handler
	Metrics *metrics.Metrics
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
	pb.RegisterProductServiceServer(s, productgrpc.NewServer(ctx.Cmd, ctx.Query))
}

func registerGin(e *gin.Engine, ctx *AppContext) {
	if ctx.Config.Server.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	api := e.Group("/api/v1")
	{
		ctx.Handler.RegisterRoutes(api)
	}
}

func initService(cfg *Config, m *metrics.Metrics) (*AppContext, func(), error) {
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default()

	// 1. 数据库
	dbWrapper, err := database.NewDB(cfg.Data.Database, cfg.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init db: %w", err)
	}
	db := dbWrapper.RawDB()
	if err := db.AutoMigrate(&productmysql.ProductModel{}, &productmysql.SKUModel{}, &productmysql.CategoryModel{}, &productmysql.BrandModel{}, &outbox.Message{}); err != nil {
		return nil, nil, fmt.Errorf("failed to migrate tables: %w", err)
	}

	// 2. 缓存
	redisCache, err := cache.NewRedisCache(&cfg.Data.Redis, cfg.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init redis: %w", err)
	}

	// 3. Search
	bootLog.Info("initializing elasticsearch client...")
	esClient, err := search.NewClient(&search.Config{
		ServiceName:         cfg.Server.Name,
		ElasticsearchConfig: cfg.Data.Elasticsearch,
		BreakerConfig:       cfg.CircuitBreaker,
		SlowThreshold:       800 * time.Millisecond,
		MaxRetries:          3,
	}, logger, m)
	if err != nil {
		bootLog.Error("failed to connect elasticsearch", "error", err)
	}

	// 4. Reliable Messaging (Outbox + Kafka)
	bootLog.Info("initializing kafka producer and outbox...")
	producer := kafka.NewProducer(&cfg.MessageQueue.Kafka, logger, m)
	outboxMgr := outbox.NewManager(db, logger.Logger)
	outboxProc := outbox.NewProcessor(outboxMgr, func(ctx context.Context, topic, key string, payload []byte) error {
		return producer.PublishToTopic(ctx, topic, []byte(key), payload)
	}, 100, 5*time.Second)
	outboxProc.Start()

	// 5. Infrastructure Repositories
	productRepo := productmysql.NewProductRepository(db)
	skuRepo := productmysql.NewSKURepository(db)
	brandRepo := productmysql.NewBrandRepository(db)
	categoryRepo := productmysql.NewCategoryRepository(db)

	productReadRepo := productredis.NewProductReadRepository(redisCache.GetClient(), cfg.Cache.DefaultExpiration)
	skuReadRepo := productredis.NewSKUReadRepository(redisCache.GetClient(), cfg.Cache.DefaultExpiration)
	brandReadRepo := productredis.NewBrandReadRepository(redisCache.GetClient(), cfg.Cache.DefaultExpiration)
	categoryReadRepo := productredis.NewCategoryReadRepository(redisCache.GetClient(), cfg.Cache.DefaultExpiration)
	searchRepo := productsearch.NewProductSearchRepository(esClient, cfg.Search.ProductIndex)

	// 6. Application Services
	publisher := outbox.NewPublisher(outboxMgr)
	cmdService := application.NewProductCommandService(
		productRepo,
		skuRepo,
		brandRepo,
		categoryRepo,
		publisher,
		logger.Logger,
	)
	queryService := application.NewProductQueryService(
		productRepo,
		skuRepo,
		brandRepo,
		categoryRepo,
		productReadRepo,
		skuReadRepo,
		brandReadRepo,
		categoryReadRepo,
		searchRepo,
		logger.Logger,
	)

	// 7. Projection Consumers (Product Events -> Read Model + ES)
	projectionService := application.NewProductProjectionService(
		productRepo,
		skuRepo,
		brandRepo,
		categoryRepo,
		productReadRepo,
		skuReadRepo,
		brandReadRepo,
		categoryReadRepo,
		searchRepo,
		logger.Logger,
	)
	projectionHandler := productconsumer.NewProductProjectionHandler(projectionService, logger.Logger)
	projectionTopics := []string{
		domain.ProductCreatedEventType,
		domain.ProductUpdatedEventType,
		domain.ProductDeletedEventType,
		domain.SKUAddedEventType,
		domain.SKUUpdatedEventType,
		domain.SKUDeletedEventType,
		domain.BrandCreatedEventType,
		domain.BrandUpdatedEventType,
		domain.BrandDeletedEventType,
		domain.CategoryCreatedEventType,
		domain.CategoryUpdatedEventType,
		domain.CategoryDeletedEventType,
	}
	projectionConsumers := make([]*kafka.Consumer, 0, len(projectionTopics))
	for _, topic := range projectionTopics {
		consumerCfg := cfg.MessageQueue.Kafka
		consumerCfg.Topic = topic
		consumerCfg.GroupID = BootstrapName + "-projection-group"
		projectionConsumer := kafka.NewConsumer(&consumerCfg, logger, m)
		projectionConsumer.Start(context.Background(), 3, projectionHandler.Handle)
		projectionConsumers = append(projectionConsumers, projectionConsumer)
	}

	httpHandler := producthttp.NewHandler(cmdService, queryService, logger.Logger)

	cleanup := func() {
		bootLog.Info("shutting down...")
		for _, c := range projectionConsumers {
			if c != nil {
				c.Close()
			}
		}
		outboxProc.Stop()
		if producer != nil {
			producer.Close()
		}
		if redisCache != nil {
			redisCache.Close()
		}
		if sqlDB, err := db.DB(); err == nil && sqlDB != nil {
			sqlDB.Close()
		}
	}

	return &AppContext{
		Config:  cfg,
		Cmd:     cmdService,
		Query:   queryService,
		Handler: httpHandler,
		Metrics: m,
	}, cleanup, nil
}

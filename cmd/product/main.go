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
	"github.com/wyfcoding/ecommerce/internal/product/infrastructure/messaging"
	"github.com/wyfcoding/ecommerce/internal/product/infrastructure/persistence"
	"github.com/wyfcoding/ecommerce/internal/product/infrastructure/persistence/elasticsearch"
	"github.com/wyfcoding/ecommerce/internal/product/interfaces/events"
	productgrpc "github.com/wyfcoding/ecommerce/internal/product/interfaces/grpc"
	producthttp "github.com/wyfcoding/ecommerce/internal/product/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/logging"
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
}

// AppContext 应用上下文
type AppContext struct {
	Config  *Config
	Cmd     *application.ProductCommandService
	Query   *application.ProductQuery
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
	db.AutoMigrate(&domain.Product{}, &domain.SKU{}, &domain.Category{}, &domain.Brand{})

	// 2. 缓存
	redisCache, err := cache.NewRedisCache(&cfg.Data.Redis, cfg.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init redis: %w", err)
	}

	// 3. Search
	esCfg := &search.Config{
		ServiceName: cfg.Server.Name,
		ElasticsearchConfig: configpkg.ElasticsearchConfig{
			Addresses: []string{"http://localhost:9200"},
		},
	}
	esClient, err := search.NewClient(esCfg, logger, m)
	if err != nil {
		bootLog.Error("failed to connect elasticsearch", "error", err)
	}

	// 4. Reliable Messaging (Outbox)
	outboxMgr := outbox.NewManager(db, logger.Logger)
	publisher := messaging.NewOutboxPublisher(outboxMgr)

	productRepo := persistence.NewProductRepository(db)
	skuRepo := persistence.NewSKURepository(db)
	brandRepo := persistence.NewBrandRepository(db)
	categoryRepo := persistence.NewCategoryRepository(db)
	searchRepo := elasticsearch.NewProductSearchRepository(esClient)

	// 6. Application Services
	cmdService := application.NewProductCommandService(
		productRepo,
		skuRepo,
		brandRepo,
		categoryRepo,
		redisCache,
		publisher,
		cfg.MessageQueue.Kafka.Topic,
		logger.Logger,
	)
	queryService := application.NewProductQuery(
		productRepo,
		skuRepo,
		brandRepo,
		categoryRepo,
		redisCache,
		logger.Logger,
		m,
		searchRepo,
	)

	// 7. Event Handlers
	searchHandler := events.NewProductSearchHandler(searchRepo)
	searchHandler.Subscribe(context.Background(), nil)

	httpHandler := producthttp.NewHandler(cmdService, queryService, logger.Logger)

	cleanup := func() {
		bootLog.Info("shutting down...")
		// Outbox manager doesn't need explicit close if DB is closed
		redisCache.Close()
	}

	return &AppContext{
		Config:  cfg,
		Cmd:     cmdService,
		Query:   queryService,
		Handler: httpHandler,
		Metrics: m,
	}, cleanup, nil
}

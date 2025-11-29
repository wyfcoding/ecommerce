package main

import (
	"context"
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/api/flashsale/v1"
	"github.com/wyfcoding/ecommerce/internal/flashsale/application"
	"github.com/wyfcoding/ecommerce/internal/flashsale/infrastructure/cache"
	"github.com/wyfcoding/ecommerce/internal/flashsale/infrastructure/persistence"
	flashsalegrpc "github.com/wyfcoding/ecommerce/internal/flashsale/interfaces/grpc"
	flashsalehttp "github.com/wyfcoding/ecommerce/internal/flashsale/interfaces/http"
	"github.com/wyfcoding/ecommerce/internal/flashsale/interfaces/mq"
	"github.com/wyfcoding/ecommerce/pkg/app"
	configpkg "github.com/wyfcoding/ecommerce/pkg/config"
	"github.com/wyfcoding/ecommerce/pkg/databases"
	"github.com/wyfcoding/ecommerce/pkg/databases/redis"
	"github.com/wyfcoding/ecommerce/pkg/idgen"
	"github.com/wyfcoding/ecommerce/pkg/logging"
	"github.com/wyfcoding/ecommerce/pkg/messagequeue/kafka"
	"github.com/wyfcoding/ecommerce/pkg/metrics"
	"github.com/wyfcoding/ecommerce/pkg/tracing"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

const serviceName = "flashsale"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9110").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.FlashSaleService)
	pb.RegisterFlashSaleServer(s, flashsalegrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for flashsale service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.FlashSaleService)
	handler := flashsalehttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for flashsale service (DDD)")
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)

	// Initialize Logger
	logger := logging.NewLogger("serviceName", "app")

	// Initialize Database
	db, err := databases.NewDB(config.Data.Database, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	// Initialize Redis Client
	redisClient, err := redis.NewRedis(&config.Data.Redis, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	// Initialize Redis Cache
	flashSaleCache := cache.NewRedisFlashSaleCache(redisClient)

	// Initialize Kafka Producer
	producer := kafka.NewProducer(config.MessageQueue.Kafka, logger)

	// Initialize ID Generator
	idGen, err := idgen.NewSnowflakeGenerator(config.Snowflake)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create id generator: %w", err)
	}

	// Infrastructure Layer
	repo := persistence.NewFlashSaleRepository(db)

	// Application Layer
	service := application.NewFlashSaleService(repo, flashSaleCache, producer, idGen, slog.Default())

	// Initialize and Start Order Consumer
	kafkaConsumer := kafka.NewConsumer(config.MessageQueue.Kafka, logger)
	orderConsumer := mq.NewOrderConsumer(kafkaConsumer, repo, logger.Logger)

	go func() {
		if err := orderConsumer.Start(context.Background()); err != nil {
			logger.Error("OrderConsumer failed", "error", err)
		}
	}()

	cleanup := func() {
		slog.Default().Info("cleaning up flashsale service resources (DDD)...")
		// Stop consumer first to stop accepting new messages
		_ = orderConsumer.Stop(context.Background())
		sqlDB.Close()
		redisClient.Close()
		producer.Close()
	}

	return service, cleanup, nil
}

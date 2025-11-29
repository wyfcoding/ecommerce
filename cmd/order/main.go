package main

import (
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/api/order/v1"
	"github.com/wyfcoding/ecommerce/internal/order/application"
	"github.com/wyfcoding/ecommerce/internal/order/infrastructure/persistence"
	ordergrpc "github.com/wyfcoding/ecommerce/internal/order/interfaces/grpc"
	orderhttp "github.com/wyfcoding/ecommerce/internal/order/interfaces/http"
	"github.com/wyfcoding/ecommerce/pkg/app"
	configpkg "github.com/wyfcoding/ecommerce/pkg/config"
	"github.com/wyfcoding/ecommerce/pkg/databases/sharding"
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
	DTM              struct {
		Server string `mapstructure:"server"`
	} `mapstructure:"dtm"`
	Warehouse struct {
		GrpcAddr string `mapstructure:"grpc_addr"`
	} `mapstructure:"warehouse"`
}

const serviceName = "order-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9093").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.OrderService)
	pb.RegisterOrderServer(s, ordergrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for order service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.OrderService)
	handler := orderhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for order service (DDD)")
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)

	// Initialize Logger
	logger := logging.NewLogger("serviceName", "app")

	// Initialize Database Sharding Manager
	shardingManager, err := sharding.NewManager(config.Data.Shards, logger)
	if err != nil {
		// Fallback to single DB if shards not configured (backward compatibility or dev mode)
		if len(config.Data.Shards) == 0 {
			logger.Info("No shards configured, using single database connection")
			// We need to wrap single DB into a manager or update repo to handle both.
			// Since repo expects manager, let's create a manager with single DB config.
			shardingManager, err = sharding.NewManager([]configpkg.DatabaseConfig{config.Data.Database}, logger)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to initialize single db manager: %w", err)
			}
		} else {
			return nil, nil, fmt.Errorf("failed to initialize sharding manager: %w", err)
		}
	}

	// Initialize ID Generator
	idGen, err := idgen.NewSnowflakeGenerator(configpkg.SnowflakeConfig{
		MachineID: 1, // NodeID 1 for order service
		StartTime: "2024-01-01",
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize id generator: %w", err)
	}

	// Initialize Kafka Producer
	producer := kafka.NewProducer(config.MessageQueue.Kafka, logger)

	// Infrastructure Layer
	repo := persistence.NewOrderRepository(shardingManager)

	// Application Layer
	service := application.NewOrderService(repo, idGen, producer, slog.Default(), config.DTM.Server, config.Warehouse.GrpcAddr, m)

	cleanup := func() {
		slog.Default().Info("cleaning up order service resources (DDD)...")
		producer.Close()
		shardingManager.Close()
	}

	return service, cleanup, nil
}

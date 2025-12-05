package main

import (
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/go-api/order_optimization/v1"
	"github.com/wyfcoding/ecommerce/internal/order_optimization/application"
	"github.com/wyfcoding/ecommerce/internal/order_optimization/infrastructure/persistence"
	optigrpc "github.com/wyfcoding/ecommerce/internal/order_optimization/interfaces/grpc"
	optihttp "github.com/wyfcoding/ecommerce/internal/order_optimization/interfaces/http"
	"github.com/wyfcoding/ecommerce/pkg/app"
	configpkg "github.com/wyfcoding/ecommerce/pkg/config"
	"github.com/wyfcoding/ecommerce/pkg/databases"
	"github.com/wyfcoding/ecommerce/pkg/logging"
	"github.com/wyfcoding/ecommerce/pkg/metrics"
	"github.com/wyfcoding/ecommerce/pkg/tracing"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

const serviceName = "order-optimization-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9122").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.OrderOptimizationService)
	pb.RegisterOrderOptimizationServiceServer(s, optigrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for order_optimization service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.OrderOptimizationService)
	handler := optihttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for order_optimization service (DDD)")
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

	// Infrastructure Layer
	repo := persistence.NewOrderOptimizationRepository(db)

	// Application Layer
	service := application.NewOrderOptimizationService(repo, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up order_optimization service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

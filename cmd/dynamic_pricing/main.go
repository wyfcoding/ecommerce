package main

import (
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/api/dynamic_pricing/v1"
	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/application"
	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/infrastructure/persistence"
	pricinggrpc "github.com/wyfcoding/ecommerce/internal/dynamic_pricing/interfaces/grpc"
	pricinghttp "github.com/wyfcoding/ecommerce/internal/dynamic_pricing/interfaces/http"
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

const serviceName = "dynamic_pricing-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9107").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.DynamicPricingService)
	pb.RegisterDynamicPricingServiceServer(s, pricinggrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for dynamic_pricing service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.DynamicPricingService)
	handler := pricinghttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for dynamic_pricing service (DDD)")
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
	repo := persistence.NewPricingRepository(db)

	// Application Layer
	service := application.NewDynamicPricingService(repo, logger.Logger)

	cleanup := func() {
		slog.Default().Info("cleaning up dynamic_pricing service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

package main

import (
	"fmt"
	"log/slog"

	"ecommerce/internal/logistics_routing/application"
	"ecommerce/internal/logistics_routing/infrastructure/persistence"
	routinghttp "ecommerce/internal/logistics_routing/interfaces/http"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	"ecommerce/pkg/databases"
	"ecommerce/pkg/logging"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/tracing"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

const serviceName = "logistics-routing-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9118").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	slog.Default().Info("gRPC server registered for logistics_routing service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.LogisticsRoutingService)
	handler := routinghttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for logistics_routing service (DDD)")
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
	repo := persistence.NewLogisticsRoutingRepository(db)

	// Application Layer
	service := application.NewLogisticsRoutingService(repo, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up logistics_routing service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

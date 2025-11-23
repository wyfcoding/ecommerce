package main

import (
	"log/slog"
	"fmt"

	"ecommerce/internal/analytics/application"
	"ecommerce/internal/analytics/infrastructure/persistence"
	analyticshttp "ecommerce/internal/analytics/interfaces/http"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	"ecommerce/pkg/databases"
	"ecommerce/pkg/idgen"
	"ecommerce/pkg/logging"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/tracing"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

const serviceName = "analytics-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9099").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	slog.Default().Info("gRPC server registered for analytics service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.AnalyticsService)
	handler := analyticshttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for analytics service (DDD)")
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

	// Initialize ID Generator
	idGenerator, err := idgen.NewSnowflakeGenerator(configpkg.SnowflakeConfig{MachineID: 1}) // NodeID should be configurable
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize id generator: %w", err)
	}

	// Infrastructure Layer
	repo := persistence.NewAnalyticsRepository(db)

	// Application Layer
	service := application.NewAnalyticsService(repo, idGenerator, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up analytics service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

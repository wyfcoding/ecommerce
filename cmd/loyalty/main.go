package main

import (
	"log/slog"
	"fmt"

	"ecommerce/internal/loyalty/application"
	"ecommerce/internal/loyalty/infrastructure/persistence"
	loyaltyhttp "ecommerce/internal/loyalty/interfaces/http"
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

const serviceName = "loyalty-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9114").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	slog.Default().Info("gRPC server registered for loyalty service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.LoyaltyService)
	handler := loyaltyhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for loyalty service (DDD)")
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
	repo := persistence.NewLoyaltyRepository(db)

	// Application Layer
	service := application.NewLoyaltyService(repo, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up loyalty service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

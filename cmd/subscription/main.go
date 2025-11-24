package main

import (
	"fmt"
	"log/slog"

	"ecommerce/internal/subscription/application"
	"ecommerce/internal/subscription/infrastructure/persistence"
	subscriptionhttp "ecommerce/internal/subscription/interfaces/http"
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

const serviceName = "subscription-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9138").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	slog.Default().Info("gRPC server registered for subscription service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.SubscriptionService)
	handler := subscriptionhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for subscription service (DDD)")
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
	repo := persistence.NewSubscriptionRepository(db)

	// Application Layer
	service := application.NewSubscriptionService(repo, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up subscription service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

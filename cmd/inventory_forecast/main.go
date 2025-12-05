package main

import (
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/go-api/inventory_forecast/v1"
	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/application"
	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/infrastructure/persistence"
	forecastgrpc "github.com/wyfcoding/ecommerce/internal/inventory_forecast/interfaces/grpc"
	forecasthttp "github.com/wyfcoding/ecommerce/internal/inventory_forecast/interfaces/http"
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

const serviceName = "inventory-forecast-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9117").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.InventoryForecastService)
	pb.RegisterInventoryForecastServiceServer(s, forecastgrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for inventory_forecast service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.InventoryForecastService)
	handler := forecasthttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for inventory_forecast service (DDD)")
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
	repo := persistence.NewInventoryForecastRepository(db)

	// Application Layer
	service := application.NewInventoryForecastService(repo, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up inventory_forecast service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

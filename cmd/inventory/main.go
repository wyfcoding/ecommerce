package main

import (
	"fmt"
	"log/slog"

	pb "ecommerce/api/inventory/v1"
	"ecommerce/internal/inventory/application"
	"ecommerce/internal/inventory/infrastructure/persistence"
	inventorygrpc "ecommerce/internal/inventory/interfaces/grpc"
	inventoryhttp "ecommerce/internal/inventory/interfaces/http"
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

const serviceName = "inventory-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9112").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.InventoryService)
	pb.RegisterInventoryServiceServer(s, inventorygrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for inventory service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.InventoryService)
	handler := inventoryhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for inventory service (DDD)")
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
	repo := persistence.NewInventoryRepository(db)
	warehouseRepo := persistence.NewWarehouseRepository(db)

	// Application Layer
	service := application.NewInventoryService(repo, warehouseRepo, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up inventory service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

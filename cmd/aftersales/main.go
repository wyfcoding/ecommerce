package main

import (
	"fmt"
	"log/slog"

	pb "ecommerce/api/aftersales/v1"
	"ecommerce/internal/aftersales/application"
	"ecommerce/internal/aftersales/infrastructure/persistence"
	aftersalesgrpc "ecommerce/internal/aftersales/interfaces/grpc"
	aftersaleshttp "ecommerce/internal/aftersales/interfaces/http"
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

const serviceName = "aftersales-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9097").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.AfterSalesService)
	pb.RegisterAftersalesServiceServer(s, aftersalesgrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for aftersales service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	aftersalesService := srv.(*application.AfterSalesService)
	handler := aftersaleshttp.NewHandler(aftersalesService, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for aftersales service (DDD)")
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

	aftersalesRepo := persistence.NewAfterSalesRepository(db)
	aftersalesService := application.NewAfterSalesService(aftersalesRepo, idGenerator, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up aftersales service resources (DDD)...")
		sqlDB.Close()
	}

	return aftersalesService, cleanup, nil
}

package main

import (
	"fmt"
	"log/slog"

	pb "ecommerce/api/data_ingestion/v1"
	"ecommerce/internal/data_ingestion/application"
	"ecommerce/internal/data_ingestion/infrastructure/persistence"
	ingestiongrpc "ecommerce/internal/data_ingestion/interfaces/grpc"
	ingestionhttp "ecommerce/internal/data_ingestion/interfaces/http"
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

const serviceName = "data_ingestion-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9105").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.DataIngestionService)
	pb.RegisterDataIngestionServer(s, ingestiongrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for data_ingestion service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.DataIngestionService)
	handler := ingestionhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for data_ingestion service (DDD)")
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
	repo := persistence.NewDataIngestionRepository(db)

	// Application Layer
	service := application.NewDataIngestionService(repo, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up data_ingestion service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

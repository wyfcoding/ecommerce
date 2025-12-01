package main

import (
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/api/data_ingestion/v1"
	"github.com/wyfcoding/ecommerce/internal/data_ingestion/application"
	"github.com/wyfcoding/ecommerce/internal/data_ingestion/infrastructure/persistence"
	ingestiongrpc "github.com/wyfcoding/ecommerce/internal/data_ingestion/interfaces/grpc"
	ingestionhttp "github.com/wyfcoding/ecommerce/internal/data_ingestion/interfaces/http"
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
	service := application.NewDataIngestionService(repo, logger.Logger)

	cleanup := func() {
		slog.Default().Info("cleaning up data_ingestion service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

package main

import (
	"log/slog"
	"fmt"

	"ecommerce/internal/ai_model/application"
	"ecommerce/internal/ai_model/infrastructure/persistence"
	aimodelhttp "ecommerce/internal/ai_model/interfaces/http"
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

const serviceName = "ai_model-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9098").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	slog.Default().Info("gRPC server registered for ai_model service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.AIModelService)
	handler := aimodelhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for ai_model service (DDD)")
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
	repo := persistence.NewAIModelRepository(db)

	// Application Layer
	service := application.NewAIModelService(repo, idGenerator, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up ai_model service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

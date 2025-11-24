package main

import (
	"fmt"
	"log/slog"

	pb "ecommerce/api/groupbuy/v1"
	"ecommerce/internal/groupbuy/application"
	"ecommerce/internal/groupbuy/infrastructure/persistence"
	groupbuygrpc "ecommerce/internal/groupbuy/interfaces/grpc"
	groupbuyhttp "ecommerce/internal/groupbuy/interfaces/http"
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

const serviceName = "groupbuy-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9110").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.GroupbuyService)
	pb.RegisterGroupbuyServiceServer(s, groupbuygrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for groupbuy service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.GroupbuyService)
	handler := groupbuyhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for groupbuy service (DDD)")
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
	repo := persistence.NewGroupbuyRepository(db)

	// ID Generator
	idGenerator, err := idgen.NewSnowflakeGenerator(configpkg.SnowflakeConfig{MachineID: 1}) // Node ID 1 for now
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create id generator: %w", err)
	}

	// Application Layer
	service := application.NewGroupbuyService(repo, idGenerator, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up groupbuy service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

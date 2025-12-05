package main

import (
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/go-api/pointsmall/v1"
	"github.com/wyfcoding/ecommerce/internal/pointsmall/application"
	"github.com/wyfcoding/ecommerce/internal/pointsmall/infrastructure/persistence"
	pointsgrpc "github.com/wyfcoding/ecommerce/internal/pointsmall/interfaces/grpc"
	pointshttp "github.com/wyfcoding/ecommerce/internal/pointsmall/interfaces/http"
	"github.com/wyfcoding/ecommerce/pkg/app"
	configpkg "github.com/wyfcoding/ecommerce/pkg/config"
	"github.com/wyfcoding/ecommerce/pkg/databases"
	"github.com/wyfcoding/ecommerce/pkg/idgen"
	"github.com/wyfcoding/ecommerce/pkg/logging"
	"github.com/wyfcoding/ecommerce/pkg/metrics"
	"github.com/wyfcoding/ecommerce/pkg/tracing"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

const serviceName = "pointsmall-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9123").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.PointsService)
	pb.RegisterPointsmallServiceServer(s, pointsgrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for pointsmall service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.PointsService)
	handler := pointshttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for pointsmall service (DDD)")
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
	idGen, err := idgen.NewSnowflakeGenerator(configpkg.SnowflakeConfig{
		MachineID: 1,
	}) // You might want to configure workerID/datacenterID
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize id generator: %w", err)
	}

	// Infrastructure Layer
	repo := persistence.NewPointsRepository(db)

	// Application Layer
	service := application.NewPointsService(repo, idGen, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up pointsmall service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

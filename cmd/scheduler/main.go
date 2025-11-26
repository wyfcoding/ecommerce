package main

import (
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/api/scheduler/v1"
	"github.com/wyfcoding/ecommerce/internal/scheduler/application"
	"github.com/wyfcoding/ecommerce/internal/scheduler/infrastructure/persistence"
	schedulergrpc "github.com/wyfcoding/ecommerce/internal/scheduler/interfaces/grpc"
	schedulerhttp "github.com/wyfcoding/ecommerce/internal/scheduler/interfaces/http"
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

const serviceName = "scheduler-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9127").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.SchedulerService)
	pb.RegisterSchedulerServiceServer(s, schedulergrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for scheduler service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.SchedulerService)
	handler := schedulerhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for scheduler service (DDD)")
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
	repo := persistence.NewSchedulerRepository(db)

	// Application Layer
	service := application.NewSchedulerService(repo, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up scheduler service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

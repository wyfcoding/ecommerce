package main

import (
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/api/review/v1"
	"github.com/wyfcoding/ecommerce/internal/review/application"
	"github.com/wyfcoding/ecommerce/internal/review/infrastructure/persistence"
	reviewgrpc "github.com/wyfcoding/ecommerce/internal/review/interfaces/grpc"
	reviewhttp "github.com/wyfcoding/ecommerce/internal/review/interfaces/http"
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

const serviceName = "review-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9092").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.ReviewService)
	pb.RegisterReviewServer(s, reviewgrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for review service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.ReviewService)
	handler := reviewhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for review service (DDD)")
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
	repo := persistence.NewReviewRepository(db)

	// Application Layer
	service := application.NewReviewService(repo, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up review service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

package main

import (
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/go-api/content_moderation/v1"
	"github.com/wyfcoding/ecommerce/internal/content_moderation/application"
	"github.com/wyfcoding/ecommerce/internal/content_moderation/infrastructure/persistence"
	moderationgrpc "github.com/wyfcoding/ecommerce/internal/content_moderation/interfaces/grpc"
	moderationhttp "github.com/wyfcoding/ecommerce/internal/content_moderation/interfaces/http"
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

const serviceName = "content_moderation-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9102").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.ModerationService)
	pb.RegisterContentModerationServer(s, moderationgrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for content_moderation service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.ModerationService)
	handler := moderationhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for content_moderation service (DDD)")
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
	repo := persistence.NewModerationRepository(db)

	// Application Layer
	service := application.NewModerationService(repo, logger.Logger)

	cleanup := func() {
		slog.Default().Info("cleaning up content_moderation service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

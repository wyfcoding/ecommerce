package main

import (
	"fmt"
	"log/slog"

	pb "ecommerce/api/search/v1"
	"ecommerce/internal/search/application"
	"ecommerce/internal/search/infrastructure/persistence"
	searchgrpc "ecommerce/internal/search/interfaces/grpc"
	searchhttp "ecommerce/internal/search/interfaces/http"
	"ecommerce/pkg/app"
	"ecommerce/pkg/cache"
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

const serviceName = "search-service"

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
	service := srv.(*application.SearchService)
	pb.RegisterSearchServiceServer(s, searchgrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for search service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.SearchService)
	handler := searchhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for search service (DDD)")
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

	// Initialize Redis
	redisCache, err := cache.NewRedisCache(config.Data.Redis)
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	// Infrastructure Layer
	repo := persistence.NewSearchRepository(db)

	// Application Layer
	service := application.NewSearchService(repo, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up search service resources (DDD)...")
		redisCache.Close()
		sqlDB.Close()
	}

	return service, cleanup, nil
}

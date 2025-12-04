package main

import (
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/api/search/v1"
	"github.com/wyfcoding/ecommerce/internal/search/application"
	"github.com/wyfcoding/ecommerce/internal/search/infrastructure/persistence"
	searchgrpc "github.com/wyfcoding/ecommerce/internal/search/interfaces/grpc"
	searchhttp "github.com/wyfcoding/ecommerce/internal/search/interfaces/http"
	"github.com/wyfcoding/ecommerce/pkg/app"
	"github.com/wyfcoding/ecommerce/pkg/cache"
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
	pb.RegisterSearchServiceServer(s, searchgrpc.NewServer(service, slog.Default()))
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

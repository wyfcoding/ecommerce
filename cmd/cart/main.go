package main

import (
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/api/cart/v1"
	"github.com/wyfcoding/ecommerce/internal/cart/application"
	"github.com/wyfcoding/ecommerce/internal/cart/infrastructure/persistence"
	cartgrpc "github.com/wyfcoding/ecommerce/internal/cart/interfaces/grpc"
	carthttp "github.com/wyfcoding/ecommerce/internal/cart/interfaces/http"
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

const serviceName = "cart-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9101").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.CartService)
	pb.RegisterCartServer(s, cartgrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for cart service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.CartService)
	handler := carthttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for cart service (DDD)")
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
	repo := persistence.NewCartRepository(db)

	// Application Layer
	service := application.NewCartService(repo, logger.Logger)

	cleanup := func() {
		slog.Default().Info("cleaning up cart service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

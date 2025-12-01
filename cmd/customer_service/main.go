package main

import (
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/api/customer_service/v1"
	"github.com/wyfcoding/ecommerce/internal/customer_service/application"
	"github.com/wyfcoding/ecommerce/internal/customer_service/infrastructure/persistence"
	csgrpc "github.com/wyfcoding/ecommerce/internal/customer_service/interfaces/grpc"
	cshttp "github.com/wyfcoding/ecommerce/internal/customer_service/interfaces/http"
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

const serviceName = "customer_service-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9104").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.CustomerService)
	pb.RegisterCustomerServiceServer(s, csgrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for customer_service service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.CustomerService)
	handler := cshttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for customer_service service (DDD)")
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
	repo := persistence.NewCustomerServiceRepository(db)

	// Application Layer
	service := application.NewCustomerService(repo, logger.Logger)

	cleanup := func() {
		slog.Default().Info("cleaning up customer_service service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

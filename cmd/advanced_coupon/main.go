package main

import (
	"fmt"
	"log/slog"

	pb "ecommerce/api/advanced_coupon/v1"
	"ecommerce/internal/advanced_coupon/application"
	"ecommerce/internal/advanced_coupon/infrastructure/persistence"
	coupongrpc "ecommerce/internal/advanced_coupon/interfaces/grpc"
	couponhttp "ecommerce/internal/advanced_coupon/interfaces/http"
	"ecommerce/pkg/app"
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

const serviceName = "advanced-coupon-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9116").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.AdvancedCouponService)
	pb.RegisterAdvancedCouponServiceServer(s, coupongrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for advanced_coupon service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.AdvancedCouponService)
	handler := couponhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for advanced_coupon service (DDD)")
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
	repo := persistence.NewAdvancedCouponRepository(db)

	// Application Layer
	service := application.NewAdvancedCouponService(repo, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up advanced_coupon service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

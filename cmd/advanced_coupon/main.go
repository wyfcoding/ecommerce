package main

import (
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/go-api/advanced_coupon/v1"
	"github.com/wyfcoding/ecommerce/internal/advanced_coupon/application"
	"github.com/wyfcoding/ecommerce/internal/advanced_coupon/infrastructure/persistence"
	coupongrpc "github.com/wyfcoding/ecommerce/internal/advanced_coupon/interfaces/grpc"
	couponhttp "github.com/wyfcoding/ecommerce/internal/advanced_coupon/interfaces/http"
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

	// 初始化 Logger
	logger := logging.NewLogger("serviceName", "app")

	// 初始化数据库
	db, err := databases.NewDB(config.Data.Database, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	// 基础设施层
	repo := persistence.NewAdvancedCouponRepository(db)

	// 应用层
	service := application.NewAdvancedCouponService(repo, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up advanced_coupon service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

package main

import (
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/go-api/aftersales/v1"
	"github.com/wyfcoding/ecommerce/internal/aftersales/application"
	"github.com/wyfcoding/ecommerce/internal/aftersales/infrastructure/persistence"
	aftersalesgrpc "github.com/wyfcoding/ecommerce/internal/aftersales/interfaces/grpc"
	aftersaleshttp "github.com/wyfcoding/ecommerce/internal/aftersales/interfaces/http"
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

const serviceName = "aftersales-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9097").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.AfterSalesService)
	pb.RegisterAftersalesServiceServer(s, aftersalesgrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for aftersales service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	aftersalesService := srv.(*application.AfterSalesService)
	handler := aftersaleshttp.NewHandler(aftersalesService, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for aftersales service (DDD)")
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

	// 初始化 ID 生成器
	idGenerator, err := idgen.NewSnowflakeGenerator(configpkg.SnowflakeConfig{MachineID: 1}) // NodeID 应该是可配置的
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize id generator: %w", err)
	}

	aftersalesRepo := persistence.NewAfterSalesRepository(db)
	aftersalesService := application.NewAfterSalesService(aftersalesRepo, idGenerator, logger.Logger)

	cleanup := func() {
		slog.Default().Info("cleaning up aftersales service resources (DDD)...")
		sqlDB.Close()
	}

	return aftersalesService, cleanup, nil
}

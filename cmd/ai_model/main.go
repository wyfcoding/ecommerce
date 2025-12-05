package main

import (
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/go-api/ai_model/v1"
	"github.com/wyfcoding/ecommerce/internal/ai_model/application"
	"github.com/wyfcoding/ecommerce/internal/ai_model/infrastructure/persistence"
	aimodelgrpc "github.com/wyfcoding/ecommerce/internal/ai_model/interfaces/grpc"
	aimodelhttp "github.com/wyfcoding/ecommerce/internal/ai_model/interfaces/http"
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

const serviceName = "ai_model-service"

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
	service := srv.(*application.AIModelService)
	pb.RegisterAIModelServiceServer(s, aimodelgrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for ai_model service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.AIModelService)
	handler := aimodelhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for ai_model service (DDD)")
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

	// 基础设施层
	repo := persistence.NewAIModelRepository(db)

	// 应用层
	service := application.NewAIModelService(repo, idGenerator, logger.Logger)

	cleanup := func() {
		slog.Default().Info("cleaning up ai_model service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

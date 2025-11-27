package main

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	pb "github.com/wyfcoding/ecommerce/api/permission/v1"
	"github.com/wyfcoding/ecommerce/internal/permission/application"
	"github.com/wyfcoding/ecommerce/internal/permission/infrastructure/persistence/mysql"
	permissiongrpc "github.com/wyfcoding/ecommerce/internal/permission/interfaces/grpc"
	"github.com/wyfcoding/ecommerce/pkg/app"
	configpkg "github.com/wyfcoding/ecommerce/pkg/config"
	"github.com/wyfcoding/ecommerce/pkg/databases"
	"github.com/wyfcoding/ecommerce/pkg/logging"
	"github.com/wyfcoding/ecommerce/pkg/metrics"
	"github.com/wyfcoding/ecommerce/pkg/tracing"

	"google.golang.org/grpc"
)

const serviceName = "permission-service"

type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9168").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.PermissionService)
	pb.RegisterPermissionServiceServer(s, permissiongrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for permission service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	// HTTP routes not implemented yet
	slog.Default().Info("HTTP routes registered for permission service (DDD)")
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)

	// Initialize Logger
	logger := logging.NewLogger(serviceName, "app")

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
	repo := mysql.NewPermissionRepository(db)

	// Application Layer
	service := application.NewPermissionService(repo, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up permission service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

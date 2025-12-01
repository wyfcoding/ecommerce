package main

import (
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/api/audit/v1"
	"github.com/wyfcoding/ecommerce/internal/audit/application"
	"github.com/wyfcoding/ecommerce/internal/audit/infrastructure/persistence"
	auditgrpc "github.com/wyfcoding/ecommerce/internal/audit/interfaces/grpc"
	audithttp "github.com/wyfcoding/ecommerce/internal/audit/interfaces/http"
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

const serviceName = "audit-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9100").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	service := srv.(*application.AuditService)
	pb.RegisterAuditServiceServer(s, auditgrpc.NewServer(service))
	slog.Default().Info("gRPC server registered for audit service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.AuditService)
	handler := audithttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for audit service (DDD)")
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

	// Initialize ID Generator
	idGenerator, err := idgen.NewSnowflakeGenerator(configpkg.SnowflakeConfig{MachineID: 1}) // NodeID should be configurable
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize id generator: %w", err)
	}

	// Infrastructure Layer
	repo := persistence.NewAuditRepository(db)

	// Application Layer
	service := application.NewAuditService(repo, idGenerator, logger.Logger)

	cleanup := func() {
		slog.Default().Info("cleaning up audit service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

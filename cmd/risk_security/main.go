package main

import (
	"fmt"
	"log/slog"

	"ecommerce/internal/risk_security/application"
	"ecommerce/internal/risk_security/infrastructure/persistence"
	riskhttp "ecommerce/internal/risk_security/interfaces/http"
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

const serviceName = "risk-security-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9126").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	slog.Default().Info("gRPC server registered for risk_security service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.RiskService)
	handler := riskhttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for risk_security service (DDD)")
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
	repo := persistence.NewRiskRepository(db)

	// Application Layer
	service := application.NewRiskService(repo, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up risk_security service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

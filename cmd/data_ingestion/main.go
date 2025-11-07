package main

import (
	"fmt"
	"time"

	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	mysqlpkg "ecommerce/pkg/database/mysql"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/tracing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Config struct {
	configpkg.Config
}

const serviceName = "data_ingestion-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9212").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	zap.S().Info("gRPC server registered")
}

func registerGin(e *gin.Engine, srv interface{}) {
	zap.S().Info("HTTP routes registered")
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)
	cleanupTracer := func() {}

	mysqlConfig := &mysqlpkg.Config{
		DSN:             config.Data.Database.DSN,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		LogLevel:        4,
		SlowThreshold:   200 * time.Millisecond,
	}
	_, cleanupDB, err := mysqlpkg.NewGORMDB(mysqlConfig, zap.L())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	cleanup := func() {
		zap.S().Info("cleaning up resources...")
		cleanupDB()
		cleanupTracer()
	}

	return nil, cleanup, nil
}

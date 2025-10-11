package main

import (
	"fmt"
	"time"

	v1 "ecommerce/api-go/api/analytics/v1"
	"ecommerce/internal/analytics/biz"
	"ecommerce/internal/analytics/data"
	"ecommerce/internal/analytics/service"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	"ecommerce/pkg/database/clickhouse"
	"ecommerce/pkg/metrics"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	gormlogger "gorm.io/gorm/logger"
)

// Config is the service-specific configuration structure.
type Config struct {
	configpkg.Config
	Data struct {
		configpkg.DataConfig
		Database struct {
			configpkg.DatabaseConfig
			LogLevel      gormlogger.LogLevel `toml:"log_level"`
			SlowThreshold time.Duration       `toml:"slow_threshold"`
		} `toml:"database"`
		Clickhouse clickhouse.Config `toml:"clickhouse"`
	} `toml:"data"`
}

func main() {
	app.NewBuilder("analytics").
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithMetrics("9097").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	v1.RegisterAnalyticsServiceServer(s, srv.(v1.AnalyticsServiceServer))
}

func registerGin(e *gin.Engine, srv interface{}) {
	analyticsHandler := handler.NewAnalyticsHandler(srv.(*service.AnalyticsService))
	// e.g., e.POST("/v1/analytics/events", analyticsHandler.RecordEvent)
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)

	chClient, chCleanup, err := clickhouse.NewClickHouseClient(&config.Data.Clickhouse)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to clickhouse: %w", err)
	}

	dataInstance, dataCleanup := data.NewData(nil) // No GORM DB for analytics
	analyticsRepo := data.NewAnalyticsRepo(dataInstance, chClient)
	analyticsUsecase := biz.NewAnalyticsUsecase(analyticsRepo)
	analyticsService := service.NewAnalyticsService(analyticsUsecase)

	cleanup := func() {
		dataCleanup()
		chCleanup()
	}

	return analyticsService, cleanup, nil
}

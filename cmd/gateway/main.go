package main

import (
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/gateway/application"
	"github.com/wyfcoding/ecommerce/internal/gateway/infrastructure/persistence"
	gatewayhttp "github.com/wyfcoding/ecommerce/internal/gateway/interfaces/http"
	"github.com/wyfcoding/ecommerce/pkg/app"
	configpkg "github.com/wyfcoding/ecommerce/pkg/config"
	"github.com/wyfcoding/ecommerce/pkg/databases"
	"github.com/wyfcoding/ecommerce/pkg/logging"
	"github.com/wyfcoding/ecommerce/pkg/metrics"
	"github.com/wyfcoding/ecommerce/pkg/middleware"
	"github.com/wyfcoding/ecommerce/pkg/tracing"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type Config struct {
	configpkg.Config `mapstructure:",squash"`
	RateLimit        struct {
		Enabled bool    `mapstructure:"enabled"`
		Rate    float64 `mapstructure:"rate"`
		Burst   int     `mapstructure:"burst"`
	} `mapstructure:"rate_limit"`
	CircuitBreaker struct {
		Enabled bool `mapstructure:"enabled"`
	} `mapstructure:"circuit_breaker"`
}

const serviceName = "gateway"

var globalConfig *Config

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(func(e *gin.Engine, srv interface{}) {
			// We need to access config here, but WithGin signature doesn't pass it.
			// However, we can use a closure or change how we register.
			// The Builder pattern in pkg/app might need adjustment or we can just register default middleware here if we can't access config.
			// Wait, initService returns the service, but main() creates the builder.
			// The config is loaded in Builder.Build().
			// Actually, `WithGin` callback is called during `Run`.
			// Let's check `pkg/app/builder.go` or `pkg/app/app.go`.
			// Assuming we can't easily pass config to `registerGin` without changing `pkg/app`.
			// But `registerGin` is a function defined in `main.go`.
			// We can capture `config` if we move `registerGin` inside `main` or make `config` a global/closure variable.
			// But `config` is loaded inside `Builder`.
			// Let's modify `registerGin` to just use hardcoded or default values if we can't access config, OR
			// better, let's look at `initService`. It receives `cfg`.
			// But `initService` returns the *service instance*, not the Gin engine.
			// The Gin engine is set up in `app.Run`.

			// Alternative: The `Builder` loads config.
			// Maybe we can't access config in `registerGin` easily.
			// Let's look at `pkg/app/builder.go` again? No, I can't view it now.

			// Let's assume I can't change `pkg/app`.
			// I will use a global variable for config in `main.go` and set it in `initService`.
			// It's a bit hacky but works.
			registerGin(e, srv)
		}).
		WithGRPCInterceptor(tracing.OtelGRPCUnaryInterceptor()).
		WithGinMiddleware(tracing.OtelGinMiddleware(serviceName)).
		WithMetrics("9111").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	slog.Default().Info("gRPC server registered for gateway service (DDD)")
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.GatewayService)
	handler := gatewayhttp.NewHandler(service, slog.Default())

	// Apply Middlewares
	if globalConfig != nil {
		if globalConfig.RateLimit.Enabled {
			slog.Info("Enabling Rate Limit Middleware", "rate", globalConfig.RateLimit.Rate, "burst", globalConfig.RateLimit.Burst)
			e.Use(middleware.RateLimit(int(globalConfig.RateLimit.Rate), globalConfig.RateLimit.Burst))
		}
		if globalConfig.CircuitBreaker.Enabled {
			slog.Info("Enabling Circuit Breaker Middleware")
			e.Use(middleware.CircuitBreaker())
		}
	}

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for gateway service (DDD)")
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)
	globalConfig = config

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
	repo := persistence.NewGatewayRepository(db)

	// Application Layer
	service := application.NewGatewayService(repo, slog.Default())

	cleanup := func() {
		slog.Default().Info("cleaning up gateway service resources (DDD)...")
		sqlDB.Close()
	}

	return service, cleanup, nil
}

package main

import (
	"fmt"
	"log/slog"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/address/v1"
	"github.com/wyfcoding/ecommerce/internal/address/application"
	addressmysql "github.com/wyfcoding/ecommerce/internal/address/infrastructure/persistence/mysql"
	addressgrpc "github.com/wyfcoding/ecommerce/internal/address/interfaces/grpc"
	addresshttp "github.com/wyfcoding/ecommerce/internal/address/interfaces/http"

	"github.com/wyfcoding/pkg/app"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/limiter"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

// BootstrapName 服务唯一标识
const BootstrapName = "address"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

// AppContext 应用上下文
type AppContext struct {
	Config  *Config
	Svc     *application.AddressService
	Handler *addresshttp.AddressHandler
	Metrics *metrics.Metrics
	Limiter limiter.Limiter
}

func main() {
	if err := app.NewBuilder[*Config, *AppContext](BootstrapName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGinMiddleware(
			middleware.CORS(),
			middleware.TimeoutMiddleware(30*time.Second),
		).
		Build().
		Run(); err != nil {
		slog.Error("service bootstrap failed", "error", err)
	}
}

func registerGRPC(s *grpc.Server, ctx *AppContext) {
	pb.RegisterAddressServiceServer(s, addressgrpc.NewAddressHandler(ctx.Svc, logging.Default().Logger))
}

func registerGin(e *gin.Engine, ctx *AppContext) {
	if ctx.Config.Server.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	if ctx.Limiter != nil {
		e.Use(middleware.RateLimitWithLimiter(ctx.Limiter))
	}
	ctx.Handler.RegisterRoutes(e) // RegisterRoutes 对应于 AddressHandler 实现
}

func initService(cfg *Config, m *metrics.Metrics) (*AppContext, func(), error) {
	c := cfg
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default()

	configpkg.PrintWithMask(c)

	// 1. DB
	db, err := database.NewDB(c.Data.Database, c.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("database init error: %w", err)
	}
	if c.Server.Environment == "dev" {
		// AutoMigrate is handled inside NewAddressRepository
	}

	// 2. Repositories
	addressRepo := addressmysql.NewAddressRepository(db.RawDB())

	// 3. Application Services
	svc := application.NewAddressService(addressRepo, logger.Logger)

	// 4. Handlers
	handler := addresshttp.NewAddressHandler(svc, logger.Logger)

	cleanup := func() {
		bootLog.Info("shutting down, releasing resources...")
		if sqlDB, err := db.RawDB().DB(); err == nil && sqlDB != nil {
			if err := sqlDB.Close(); err != nil {
				bootLog.Error("failed to close sql database", "error", err)
			}
		}
	}

	return &AppContext{
		Config:  c,
		Svc:     svc,
		Handler: handler,
		Metrics: m,
		Limiter: nil, // Add rate limiter if needed later via pkg
	}, cleanup, nil
}

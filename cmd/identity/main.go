package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/identity/application"
	"github.com/wyfcoding/ecommerce/internal/identity/domain"
	identitymysql "github.com/wyfcoding/ecommerce/internal/identity/infrastructure/persistence/mysql"
	identityhttp "github.com/wyfcoding/ecommerce/internal/identity/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/limiter"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"

	"github.com/gin-gonic/gin"
)

const BootstrapName = "identity"

type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

type AppContext struct {
	Config     *Config
	Handler    *identityhttp.Handler
	Metrics    *metrics.Metrics
	Limiter    limiter.Limiter
}

func main() {
	if err := app.NewBuilder[*Config, *AppContext](BootstrapName).
		WithConfig(&Config{}).
		WithService(initService).
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

func registerGin(e *gin.Engine, ctx *AppContext) {
	if ctx.Config.Server.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	if ctx.Limiter != nil {
		e.Use(middleware.RateLimitWithLimiter(ctx.Limiter))
	}

	e.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := e.Group("/api/v1")
	ctx.Handler.RegisterRoutes(api)
}

func initService(cfg *Config, m *metrics.Metrics) (*AppContext, func(), error) {
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default()

	configpkg.PrintWithMask(cfg)

	db, err := database.NewDB(cfg.Data.Database, cfg.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("database init error: %w", err)
	}

	if cfg.Server.Environment == "dev" {
		if err := db.RawDB().AutoMigrate(
			&identitymysql.UserPersonaModel{},
			&identitymysql.LinkedAccountModel{},
			&identitymysql.AuthSessionModel{},
			&identitymysql.UserMappingModel{},
			&identitymysql.KYCRecordModel{},
			&identitymysql.KYCSessionModel{},
			&identitymysql.RiskProfileModel{},
		); err != nil {
			bootLog.Error("failed to migrate database", "error", err)
		}
	}

	redisCache, err := cache.NewRedisCache(&cfg.Data.Redis, cfg.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("redis init error: %w", err)
	}

	kycRepo := identitymysql.NewKYCRepository(db.RawDB())

	bridge := domain.NewIdentityBridgeService()
	identitySvc := application.NewIdentityApplicationService(kycRepo, bridge, &mockTradingAuthClient{})

	kycBridge := application.NewKYCBridgeService(kycRepo, &mockComplianceClient{})

	handler := identityhttp.NewHandler(identitySvc, kycBridge)

	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), cfg.RateLimit.Rate, cfg.RateLimit.Burst)

	cleanup := func() {
		bootLog.Info("shutting down, releasing resources...")
		if redisCache != nil {
			if err := redisCache.Close(); err != nil {
				bootLog.Error("failed to close redis cache", "error", err)
			}
		}
		if sqlDB, err := db.RawDB().DB(); err == nil && sqlDB != nil {
			if err := sqlDB.Close(); err != nil {
				bootLog.Error("failed to close sql database", "error", err)
			}
		}
	}

	return &AppContext{
		Config:  cfg,
		Handler: handler,
		Metrics: m,
		Limiter: rateLimiter,
	}, cleanup, nil
}

type mockTradingAuthClient struct{}

func (m *mockTradingAuthClient) VerifyToken(ctx context.Context, token string) (string, error) {
	return "TRADING-USER-123", nil
}

func (m *mockTradingAuthClient) GenerateToken(ctx context.Context, userID string) (string, error) {
	return "TRADING-TOKEN-" + userID, nil
}

type mockComplianceClient struct{}

func (m *mockComplianceClient) SyncKYCStatus(ctx context.Context, userID string, level int, status string) error {
	return nil
}

func (m *mockComplianceClient) CheckAML(ctx context.Context, userID string, idNumber string) (bool, string, error) {
	return true, "", nil
}

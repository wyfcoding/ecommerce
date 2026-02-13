package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/riskanalyzer/application"
	"github.com/wyfcoding/ecommerce/internal/riskanalyzer/domain"
	riskmysql "github.com/wyfcoding/ecommerce/internal/riskanalyzer/infrastructure/persistence/mysql"
	riskhttp "github.com/wyfcoding/ecommerce/internal/riskanalyzer/interfaces/http"
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

const BootstrapName = "riskanalyzer"

type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

type AppContext struct {
	Config  *Config
	Handler *riskhttp.Handler
	Metrics *metrics.Metrics
	Limiter limiter.Limiter
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
			&riskmysql.RiskAssessmentModel{},
			&riskmysql.RiskRuleModel{},
			&riskmysql.BlacklistModel{},
			&riskmysql.CreditProfileModel{},
			&riskmysql.CreditRecordModel{},
		); err != nil {
			bootLog.Error("failed to migrate database", "error", err)
		}
	}

	redisCache, err := cache.NewRedisCache(&cfg.Data.Redis, cfg.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("redis init error: %w", err)
	}

	riskRepo := riskmysql.NewRiskRepository(db.RawDB())
	creditRepo := riskmysql.NewCreditProfileRepository(db.RawDB())

	fraudEngine := &defaultFraudEngine{}
	creditEval := &defaultCreditEvaluator{}
	riskBridge := &defaultRiskBridge{}

	cmdSvc := application.NewRiskCommandService(
		riskRepo,
		creditRepo,
		fraudEngine,
		creditEval,
		riskBridge,
		nil,
		"risk.events",
		logger.Logger,
	)

	querySvc := application.NewRiskQueryService(
		riskRepo,
		creditRepo,
		logger.Logger,
	)

	handler := riskhttp.NewHandler(cmdSvc, querySvc)

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

type defaultFraudEngine struct{}

func (e *defaultFraudEngine) Analyze(ctx context.Context, assessment *domain.RiskAssessment) (*domain.RiskAssessment, error) {
	assessment.TotalScore = 10
	assessment.Level = domain.RiskLow
	assessment.Decision = "PASS"
	return assessment, nil
}

func (e *defaultFraudEngine) GetDeviceTrustScore(ctx context.Context, deviceFinger string) (int, error) {
	return 80, nil
}

type defaultCreditEvaluator struct{}

func (e *defaultCreditEvaluator) Evaluate(ctx context.Context, userID uint64) (*domain.CreditProfile, error) {
	return &domain.CreditProfile{
		UserID:         userID,
		Score:          650,
		Level:          domain.LevelGood,
		TotalLimit:     100000,
		UsedLimit:      0,
		AvailableLimit: 100000,
		Status:         "ACTIVE",
		LastAssessedAt: time.Now(),
	}, nil
}

func (e *defaultCreditEvaluator) CalculateLimit(ctx context.Context, profile *domain.CreditProfile) (int64, error) {
	return profile.TotalLimit, nil
}

type defaultRiskBridge struct{}

func (b *defaultRiskBridge) ReportSignal(ctx context.Context, signal *domain.SharedRiskSignal) error {
	return nil
}

func (b *defaultRiskBridge) GetUnifiedRiskScore(ctx context.Context, userID string) (float64, error) {
	return 0, nil
}

func (b *defaultRiskBridge) ApplyRestriction(ctx context.Context, userID string, level domain.RiskLevel) error {
	return nil
}

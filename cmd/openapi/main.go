package main

import (
	"fmt"
	"log/slog"

	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/go-api/openapi/v1"
	"github.com/wyfcoding/ecommerce/internal/openapi/application"
	"github.com/wyfcoding/ecommerce/internal/openapi/domain"
	"github.com/wyfcoding/ecommerce/internal/openapi/infrastructure"
	"github.com/wyfcoding/ecommerce/internal/openapi/interfaces"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
)

// BootstrapName 服务唯一标识
const BootstrapName = "openapi"

// Config 服务扩展配置
type Config struct {
	config.Config `mapstructure:",squash"`
}

// AppContext 应用上下文
type AppContext struct {
	Config     *Config
	AppService *application.OpenApiService
	Metrics    *metrics.Metrics
}

func main() {
	if err := app.NewBuilder[*Config, *AppContext](BootstrapName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		Build().
		Run(); err != nil {
		slog.Error("service bootstrap failed", "error", err)
	}
}

func registerGRPC(s *grpc.Server, ctx *AppContext) {
	pb.RegisterOpenApiServiceServer(s, interfaces.NewOpenApiHandler(ctx.AppService))
}

func initService(cfg *Config, m *metrics.Metrics) (*AppContext, func(), error) {
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default()

	// 1. 数据库
	dbWrapper, err := database.NewDB(cfg.Data.Database, cfg.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init db: %w", err)
	}
	db := dbWrapper.RawDB()

	// 自动迁移
	if err := db.AutoMigrate(&domain.OpenApiApp{}); err != nil {
		return nil, nil, fmt.Errorf("failed to migrate tables: %w", err)
	}

	// 2. 依赖注入
	repo := infrastructure.NewGormOpenApiRepository(db)
	appService := application.NewOpenApiService(repo)

	cleanup := func() {
		bootLog.Info("shutting down...")
		if sqlDB, err := db.DB(); err == nil && sqlDB != nil {
			sqlDB.Close()
		}
	}

	return &AppContext{
		Config:     cfg,
		AppService: appService,
		Metrics:    m,
	}, cleanup, nil
}

// 注释说明：OpenApi 服务目前已支持标准的 app.Builder 范式。
// 数据库 DSN 将从 configs/openapi/config.toml 中加载并通过环境变量覆盖。

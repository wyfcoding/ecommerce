// 生成摘要：
// - 实现 Wallet 服务的启动入口，采用封装好的 app.Builder 进行标准化引导
// - 完成依赖注入：MySQL DB -> Repository -> Application Service -> gRPC Handler
// - 支持自动数据库迁移、优雅关停、日志/指标/追踪集成
// - 遵循 DDD 与微服务架构规范

package main

import (
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/go-api/wallet/v1"
	"github.com/wyfcoding/ecommerce/internal/wallet/application"
	walletmysql "github.com/wyfcoding/ecommerce/internal/wallet/infrastructure/persistence/mysql"
	walletgrpc "github.com/wyfcoding/ecommerce/internal/wallet/interfaces/grpc"
	"github.com/wyfcoding/pkg/app"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"google.golang.org/grpc"
)

// BootstrapName 服务唯一标识
const BootstrapName = "wallet"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

// AppContext 应用上下文，承载核心依赖
type AppContext struct {
	Config  *Config
	Service *application.WalletService
}

func main() {
	if err := app.NewBuilder[*Config, *AppContext](BootstrapName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		Build().
		Run(); err != nil {
		slog.Error("wallet service bootstrap failed", "error", err)
	}
}

// registerGRPC 注册 gRPC 服务实现
func registerGRPC(s *grpc.Server, ctx *AppContext) {
	pb.RegisterWalletServiceServer(s, walletgrpc.NewServer(ctx.Service, slog.Default()))
}

// initService 核心依赖初始化
func initService(cfg *Config, m *metrics.Metrics) (*AppContext, func(), error) {
	c := cfg
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default()

	configpkg.PrintWithMask(c)

	// 1. 初始化数据库连接
	db, err := database.NewDB(c.Data.Database, c.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("database init error: %w", err)
	}

	// 2. 自动迁移 (Dev 环境)
	if c.Server.Environment == "dev" {
		bootLog.Info("running auto-migration for wallet service")
		if err := db.RawDB().AutoMigrate(
			&walletmysql.WalletModel{},
			&walletmysql.TransactionModel{},
			&walletmysql.WalletLimitModel{},
			&walletmysql.DailyUsageModel{},
		); err != nil {
			bootLog.Error("failed to migrate database", "error", err)
		}
	}

	// 3. 构建仓储层与应用服务
	walletRepo := walletmysql.NewWalletRepository(db, logger.Logger)
	txRepo := walletmysql.NewTransactionRepository(db, logger.Logger)

	walletSvc := application.NewWalletService(walletRepo, txRepo, logger.Logger)

	// 资源清理回调
	cleanup := func() {
		bootLog.Info("shutting down wallet service, releasing resources...")
		if sqlDB, err := db.RawDB().DB(); err == nil && sqlDB != nil {
			sqlDB.Close()
		}
	}

	return &AppContext{
		Config:  c,
		Service: walletSvc,
	}, cleanup, nil
}

package main

import (
	"fmt"
	"log/slog"

	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/go-api/orchestrator/v1"
	"github.com/wyfcoding/ecommerce/internal/orchestrator/application"
	"github.com/wyfcoding/ecommerce/internal/orchestrator/infrastructure/grpcclient"
	"github.com/wyfcoding/ecommerce/internal/orchestrator/infrastructure/persistence/mysql"
	"github.com/wyfcoding/ecommerce/internal/orchestrator/interfaces"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database"
	pkggrpc "github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/saga"
)

// BootstrapName 服务唯一标识
const BootstrapName = "orchestrator"

// Config 服务扩展配置
type Config struct {
	config.Config `mapstructure:",squash"`
}

// AppContext 应用上下文
type AppContext struct {
	Config     *Config
	AppService *application.OrchestratorApplicationService
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
	pb.RegisterOrchestratorServiceServer(s, interfaces.NewOrchestratorHandler(ctx.AppService))
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
	// if err := db.AutoMigrate(&domain.SagaInstance{}, &domain.SagaStep{}); err != nil {
	// 	return nil, nil, fmt.Errorf("failed to migrate tables: %w", err)
	// }

	// 2. 初始化 gRPC 客户端
	sClients := &grpcclient.ServiceClients{}
	grpcCleanup, err := pkggrpc.InitClients(cfg.Services, m, cfg.CircuitBreaker, sClients)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init grpc clients: %w", err)
	}
	sClients.Init()

	// 3. 依赖注入
	repo := mysql.NewRepository(db)
	engine := saga.NewEngine()
	appService := application.NewOrchestratorApplicationService(repo, engine, logger.Logger, sClients)

	cleanup := func() {
		bootLog.Info("shutting down...")
		grpcCleanup()
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

// 注释说明：AutoMigrate 已被禁用，建议通过迁移脚本管理。
// 若需开启，请取消 core domain 的注释。目前因 domain 与 PB 类型可能冲突，建议手动管理表结构或谨慎迁移。

package main

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/goapi/admin/v1"
	"github.com/wyfcoding/ecommerce/internal/admin/application"
	"github.com/wyfcoding/ecommerce/internal/admin/infrastructure/persistence/mysql"
	admingrpc "github.com/wyfcoding/ecommerce/internal/admin/interfaces/grpc"
	adminhttp "github.com/wyfcoding/ecommerce/internal/admin/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
)

// BootstrapName 服务名称常量。
const BootstrapName = "admin"

// AppContext 应用上下文，包含配置、服务实例和客户端依赖。
type AppContext struct {
	Config          *configpkg.Config
	Admin           *application.AdminService
	Clients         *ServiceClients
	AuthHandler     *adminhttp.AuthHandler
	WorkflowHandler *adminhttp.WorkflowHandler
}

// ServiceClients 包含所有下游服务的 gRPC 客户端连接。
type ServiceClients struct {
	User    *grpc.ClientConn
	Order   *grpc.ClientConn
	Payment *grpc.ClientConn
}

func main() {
	if err := app.NewBuilder(BootstrapName).
		WithConfig(&configpkg.Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGinMiddleware(middleware.CORS()).
		Build().
		Run(); err != nil {
		slog.Error("application run failed", "error", err)
	}
}

func registerGRPC(s *grpc.Server, svc any) {
	ctx := svc.(*AppContext)
	pb.RegisterAdminServer(s, admingrpc.NewServer(ctx.Admin))
}

func registerGin(e *gin.Engine, svc any) {
	ctx := svc.(*AppContext)
	api := e.Group("/api/v1")

	ctx.AuthHandler.RegisterRoutes(api)
	ctx.WorkflowHandler.RegisterRoutes(api)
}

// 重写 registerGin 以使用带有 Handlers 的 AppContext
/*
// AppContext 应用上下文，包含配置、服务实例和客户端依赖。
type AppContext struct {
    ...
    AuthHandler     *adminhttp.AuthHandler
    WorkflowHandler *adminhttp.WorkflowHandler
}
*/

func initService(cfg any, m *metrics.Metrics) (any, func(), error) {
	c := cfg.(*configpkg.Config)
	slog.Info("initializing service dependencies...")

	// 1. 数据库
	db, err := databases.NewDB(c.Data.Database, logging.Default())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 2. Redis 缓存
	// 暂时禁用，除非需要避免未使用变量错误
	_, err = cache.NewRedisCache(c.Data.Redis)
	if err != nil {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	// 3. Service Clients
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, clients)
	if err != nil {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 4. Repositories (from mysql package)
	adminRepo := mysql.NewAdminRepository(db)
	roleRepo := mysql.NewRoleRepository(db)
	auditRepo := mysql.NewAuditRepository(db)
	approvalRepo := mysql.NewApprovalRepository(db)
	settingRepo := mysql.NewSettingRepository(db)

	// 5. Dependencies
	logger := slog.Default()
	opsDeps := application.SystemOpsDependencies{
		OrderClient:   clients.Order,
		UserClient:    clients.User,
		PaymentClient: clients.Payment,
	}

	// 6. Application Facade
	adminService := application.NewAdminService(
		adminRepo,
		roleRepo,
		auditRepo,
		settingRepo,
		approvalRepo,
		opsDeps,
		logger,
	)

	// 7. Handlers
	authHandler := adminhttp.NewAuthHandler(adminService, logger)
	workflowHandler := adminhttp.NewWorkflowHandler(adminService, logger)

	cleanup := func() {
		slog.Info("cleaning up resources...")
		clientCleanup()
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	return &AppContext{
		Config:          c,
		Admin:           adminService,
		Clients:         clients,
		AuthHandler:     authHandler,
		WorkflowHandler: workflowHandler,
	}, cleanup, nil
}

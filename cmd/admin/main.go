package main

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/go-api/admin/v1"
	"github.com/wyfcoding/ecommerce/internal/admin/application"
	"github.com/wyfcoding/ecommerce/internal/admin/infrastructure/persistence"
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
	AdminService    *application.AdminService
	Clients         *ServiceClients
	AuthHandler     *adminhttp.AuthHandler
	WorkflowHandler *adminhttp.WorkflowHandler
}

// ServiceClients 包含所有下游服务的 gRPC 客户端连接。
type ServiceClients struct {
	User         *grpc.ClientConn
	Product      *grpc.ClientConn
	Order        *grpc.ClientConn
	Cart         *grpc.ClientConn
	Payment      *grpc.ClientConn
	Inventory    *grpc.ClientConn
	Notification *grpc.ClientConn
	Logistics    *grpc.ClientConn
	Coupon       *grpc.ClientConn
	Review       *grpc.ClientConn
	Wishlist     *grpc.ClientConn
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
	pb.RegisterAdminServiceServer(s, admingrpc.NewServer(ctx.AdminService))
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
	// redisCache, err := cache.NewRedisCache(c.Data.Redis) // 暂未使用，但保留备用
	// 暂时禁用，除非需要避免未使用变量错误
	_, err = cache.NewRedisCache(c.Data.Redis)
	if err != nil {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	// 3. Service Clients
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitServiceClients(c.Services, clients)
	if err != nil {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 4. Repositories
	adminRepo := persistence.NewAdminRepository(db)
	roleRepo := persistence.NewRoleRepository(db) // 修复：现在已存在
	auditRepo := persistence.NewAuditRepository(db)
	approvalRepo := persistence.NewApprovalRepository(db)
	settingRepo := persistence.NewSettingRepository(db)

	// 5. Domain Services
	logger := slog.Default()
	authService := application.NewAdminAuthService(adminRepo, roleRepo, logger)
	auditService := application.NewAuditService(auditRepo, logger)

	opsDeps := application.SystemOpsDependencies{
		OrderClient:   clients.Order,
		UserClient:    clients.User,
		PaymentClient: clients.Payment,
	}
	opsService := application.NewSystemOpsService(opsDeps, logger)

	workflowService := application.NewWorkflowService(approvalRepo, opsService, auditService, logger)

	// 6. Application Facade
	adminService := application.NewAdminService(
		adminRepo,
		roleRepo,
		auditRepo,
		settingRepo,
		authService,
		auditService,
		workflowService,
		logger,
	)

	// 7. Handlers
	authHandler := adminhttp.NewAuthHandler(authService, logger)
	workflowHandler := adminhttp.NewWorkflowHandler(workflowService, logger)

	cleanup := func() {
		slog.Info("cleaning up resources...")
		clientCleanup()
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	return &AppContext{
		Config:          c,
		AdminService:    adminService,
		Clients:         clients,
		AuthHandler:     authHandler,
		WorkflowHandler: workflowHandler,
	}, cleanup, nil
}

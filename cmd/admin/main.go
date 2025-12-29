package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

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
	"github.com/wyfcoding/pkg/idempotency"
	"github.com/wyfcoding/pkg/limiter"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
)

// BootstrapName 服务标识。
const BootstrapName = "admin"

// Config 扩展配置结构。
type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

// AppContext 应用资源上下文。
type AppContext struct {
	Config          *Config
	Admin           *application.AdminService
	Clients         *ServiceClients
	AuthHandler     *adminhttp.AuthHandler
	WorkflowHandler *adminhttp.WorkflowHandler
	Metrics         *metrics.Metrics
	Limiter         limiter.Limiter
	Idempotency     idempotency.Manager
}

// ServiceClients 外部 gRPC 服务客户端连接池。
type ServiceClients struct {
	User    *grpc.ClientConn `service:"user"`
	Order   *grpc.ClientConn `service:"order"`
	Payment *grpc.ClientConn `service:"payment"`
}

func main() {
	if err := app.NewBuilder(BootstrapName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGinMiddleware(
			middleware.MetricsMiddleware(), // Prometheus 指标收集中间件
			middleware.CORS(),              // 跨域支持
		).
		Build().
		Run(); err != nil {
		slog.Error("service bootstrap failed", "error", err)
	}
}

// registerGRPC 注册 gRPC 服务。
func registerGRPC(s *grpc.Server, svc any) {
	ctx := svc.(*AppContext)
	pb.RegisterAdminServer(s, admingrpc.NewServer(ctx.Admin))
}

// registerGin 注册 HTTP 路由。
func registerGin(e *gin.Engine, svc any) {
	ctx := svc.(*AppContext)

	// 1. 运行模式预设 (必须在所有路由注册前)
	if ctx.Config.Server.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 2. 基础路由层：不应用限流和幂等
	sys := e.Group("/sys")
	{
		sys.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "UP",
				"service":   BootstrapName,
				"timestamp": time.Now().Unix(),
			})
		})
		sys.GET("/ready", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "READY"})
		})
	}

	// 监控指标端点
	if ctx.Config.Metrics.Enabled {
		e.GET(ctx.Config.Metrics.Path, gin.WrapH(ctx.Metrics.Handler()))
	}

	// 3. 业务保护层：限流
	e.Use(middleware.RateLimitWithLimiter(ctx.Limiter))

	// 4. 业务逻辑路由组：加入幂等保护
	api := e.Group("/api/v1")
	{
		// 仅针对写操作接口开启幂等保护
		api.Use(middleware.IdempotencyMiddleware(ctx.Idempotency, 24*time.Hour))
		
		ctx.AuthHandler.RegisterRoutes(api)
		ctx.WorkflowHandler.RegisterRoutes(api)
	}

	slog.Info("HTTP service configured successfully", "service", BootstrapName)
}

func initService(cfg any, m *metrics.Metrics) (any, func(), error) {
	c := cfg.(*Config)
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default()

	// 1. 基础设施初始化
	db, err := databases.NewDB(c.Data.Database, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init database: %w", err)
	}

	redisCache, err := cache.NewRedisCache(c.Data.Redis)
	if err != nil {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("failed to init redis: %w", err)
	}

	// 2. 治理：分布式限流器 (基于 Redis)
	rateLimiter := limiter.NewRedisLimiter(
		redisCache.GetClient(),
		c.RateLimit.Rate,
		time.Second,
	)

	// 3. 治理：分布式幂等管理器 (基于 Redis)
	idemManager := idempotency.NewRedisManager(redisCache.GetClient(), "admin:idem")

	// 4. 外部服务连接 (自动发现)
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, clients)
	if err != nil {
		redisCache.Close()
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("failed to init grpc clients: %w", err)
	}

	// 4. 业务分层装配 (DDD 结构)
	bootLog.Info("assembling domain services...")

	adminRepo := mysql.NewAdminRepository(db)
	roleRepo := mysql.NewRoleRepository(db)
	auditRepo := mysql.NewAuditRepository(db)
	approvalRepo := mysql.NewApprovalRepository(db)
	settingRepo := mysql.NewSettingRepository(db)

	opsDeps := application.SystemOpsDependencies{
		OrderClient:   clients.Order,
		UserClient:    clients.User,
		PaymentClient: clients.Payment,
	}

	adminService := application.NewAdminService(
		adminRepo,
		roleRepo,
		auditRepo,
		settingRepo,
		approvalRepo,
		opsDeps,
		logger.Logger,
	)

	authHandler := adminhttp.NewAuthHandler(adminService, logger.Logger)
	workflowHandler := adminhttp.NewWorkflowHandler(adminService, logger.Logger)

	// 5. 资源回收逻辑
	cleanup := func() {
		bootLog.Info("shutting down, releasing resources...")
		clientCleanup()
		if redisCache != nil {
			redisCache.Close()
		}
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}

	return &AppContext{
		Config:          c,
		Admin:           adminService,
		Clients:         clients,
		AuthHandler:     authHandler,
		WorkflowHandler: workflowHandler,
		Metrics:         m,
		Limiter:         rateLimiter,
		Idempotency:     idemManager,
	}, cleanup, nil
}

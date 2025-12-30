package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/pkg/response"

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
	"github.com/wyfcoding/pkg/storage"
)

// BootstrapName 服务唯一标识
const BootstrapName = "admin"

// IdempotencyPrefix 幂等性 Redis 键前缀
const IdempotencyPrefix = "admin:idem"

// Config 服务扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

// AppContext 应用上下文 (包含对外服务实例与依赖)
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

// ServiceClients 下游微服务客户端集合
type ServiceClients struct {
	User    *grpc.ClientConn `service:"user"`
	Order   *grpc.ClientConn `service:"order"`
	Payment *grpc.ClientConn `service:"payment"`
}

func main() {
	// 构建并运行服务
	if err := app.NewBuilder(BootstrapName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGinMiddleware(
			middleware.MetricsMiddleware(),               // 指标采集
			middleware.CORS(),                            // 跨域处理
			middleware.TimeoutMiddleware(30*time.Second), // 全局超时 (注: 此处无法读取配置，使用默认值)
		).
		Build().
		Run(); err != nil {
		slog.Error("service bootstrap failed", "error", err)
	}
}

// registerGRPC 注册 gRPC 服务
func registerGRPC(s *grpc.Server, svc any) {
	ctx := svc.(*AppContext)
	pb.RegisterAdminServiceServer(s, admingrpc.NewServer(ctx.Admin))
}

// registerGin 注册 HTTP 路由
func registerGin(e *gin.Engine, svc any) {
	ctx := svc.(*AppContext)

	// 根据环境设置 Gin 模式
	if ctx.Config.Server.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 系统检查接口
	sys := e.Group("/sys")
	{
		sys.GET("/health", func(c *gin.Context) {
			response.SuccessWithRawData(c, gin.H{
				"status":    "UP",
				"service":   BootstrapName,
				"timestamp": time.Now().Unix(),
			})
		})
		sys.GET("/ready", func(c *gin.Context) {
			response.SuccessWithRawData(c, gin.H{"status": "READY"})
		})
	}

	// 指标暴露
	if ctx.Config.Metrics.Enabled {
		e.GET(ctx.Config.Metrics.Path, gin.WrapH(ctx.Metrics.Handler()))
	}

	// 全局限流中间件
	e.Use(middleware.RateLimitWithLimiter(ctx.Limiter))

	// 业务 API 路由 v1
	api := e.Group("/api/v1")
	{
		// 公开接口 (如登录)
		ctx.AuthHandler.RegisterRoutes(api, ctx.Config.JWT.Secret)

		// 鉴权接口
		protected := api.Group("/")
		protected.Use(middleware.JWTAuth(ctx.Config.JWT.Secret))
		protected.Use(middleware.IdempotencyMiddleware(ctx.Idempotency, 24*time.Hour))
		{
			ctx.WorkflowHandler.RegisterRoutes(protected)
		}
	}
}

// initService 初始化服务依赖 (数据库、缓存、客户端、领域层)
func initService(cfg any, m *metrics.Metrics) (any, func(), error) {
	c := cfg.(*Config)
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default() // 获取全局 Logger

	// 打印脱敏配置
	configpkg.PrintWithMask(c)

	// 1. 初始化数据库 (MySQL)
	db, err := databases.NewDB(c.Data.Database, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("database init error: %w", err)
	}

	// 2. 初始化缓存 (Redis)
	redisCache, err := cache.NewRedisCache(c.Data.Redis, logger)
	if err != nil {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("redis init error: %w", err)
	}

	// 3. 初始化治理组件 (限流器、幂等管理器)
	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), c.RateLimit.Rate, time.Second)
	idemManager := idempotency.NewRedisManager(redisCache.GetClient(), IdempotencyPrefix)

	// 4. 初始化存储基础设施 (MinIO)
	// 注意: 作为通用能力注入到 Service，而非作为 Repository
	store, err := storage.NewMinIOClient(
		c.Minio.Endpoint,
		c.Minio.AccessKeyID,
		c.Minio.SecretAccessKey,
		c.Minio.BucketName,
		c.Minio.UseSSL,
	)
	if err != nil {
		bootLog.Warn("storage init failed, continuing without storage", "error", err)
	}

	// 5. 初始化下游微服务客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, clients)
	if err != nil {
		redisCache.Close()
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("grpc clients init error: %w", err)
	}

	// 6. DDD 分层装配 (Infrastructure -> Domain -> Application -> Interface)
	bootLog.Info("assembling services with full dependency injection...")

	// 6.1 Infrastructure (Persistence)
	adminRepo := mysql.NewAdminRepository(db)
	roleRepo := mysql.NewRoleRepository(db)
	auditRepo := mysql.NewAuditRepository(db)
	approvalRepo := mysql.NewApprovalRepository(db)
	settingRepo := mysql.NewSettingRepository(db)

	// 6.2 Application (Service)
	// 注入外部依赖 (Parameter Object Pattern)
	opsDeps := application.SystemOpsDependencies{
		OrderClient:   clients.Order,
		UserClient:    clients.User,
		PaymentClient: clients.Payment,
		Storage:       store,
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

	// 6.3 Interface (HTTP Handlers)
	authHandler := adminhttp.NewAuthHandler(adminService, logger.Logger)
	workflowHandler := adminhttp.NewWorkflowHandler(adminService, logger.Logger)

	// 定义资源清理函数
	cleanup := func() {
		bootLog.Info("shutting down, releasing resources...")
		clientCleanup()
		if redisCache != nil {
			if err := redisCache.Close(); err != nil {
				bootLog.Error("failed to close redis cache", "error", err)
			}
		}
		if sqlDB, err := db.DB(); err == nil && sqlDB != nil {
			if err := sqlDB.Close(); err != nil {
				bootLog.Error("failed to close sql database", "error", err)
			}
		}
	}

	// 返回应用上下文与清理函数
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

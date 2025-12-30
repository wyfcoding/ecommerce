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
	"github.com/wyfcoding/pkg/storage"
)

// BootstrapName 服务标识
const BootstrapName = "admin"

// Config 扩展配置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

// AppContext 应用资源上下文 (仅包含对外提供服务的实例)
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

// ServiceClients 下游微服务
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
			middleware.MetricsMiddleware(),
			middleware.CORS(),
			middleware.TimeoutMiddleware(30*time.Second),
		).
		Build().
		Run(); err != nil {
		slog.Error("service bootstrap failed", "error", err)
	}
}

func registerGRPC(s *grpc.Server, svc any) {
	ctx := svc.(*AppContext)
	pb.RegisterAdminServer(s, admingrpc.NewServer(ctx.Admin))
}

func registerGin(e *gin.Engine, svc any) {
	ctx := svc.(*AppContext)

	if ctx.Config.Server.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

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

	if ctx.Config.Metrics.Enabled {
		e.GET(ctx.Config.Metrics.Path, gin.WrapH(ctx.Metrics.Handler()))
	}

	e.Use(middleware.RateLimitWithLimiter(ctx.Limiter))

	api := e.Group("/api/v1")
	{
		ctx.AuthHandler.RegisterRoutes(api, ctx.Config.JWT.Secret)

		protected := api.Group("/")
		protected.Use(middleware.JWTAuth(ctx.Config.JWT.Secret))
		protected.Use(middleware.IdempotencyMiddleware(ctx.Idempotency, 24*time.Hour))
		{
			ctx.WorkflowHandler.RegisterRoutes(protected)
		}
	}
}

func initService(cfg any, m *metrics.Metrics) (any, func(), error) {
	c := cfg.(*Config)
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default()

	configpkg.PrintWithMask(c)

	// 1. 数据库
	db, err := databases.NewDB(c.Data.Database, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("database init error: %w", err)
	}

	// 2. Redis
	redisCache, err := cache.NewRedisCache(c.Data.Redis)
	if err != nil {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("redis init error: %w", err)
	}

	// 3. 治理组件
	rateLimiter := limiter.NewRedisLimiter(redisCache.GetClient(), c.RateLimit.Rate, time.Second)
	idemManager := idempotency.NewRedisManager(redisCache.GetClient(), "admin:idem")

	// 4. 存储基础设施 (不再放入 AppContext，而是准备注入 Service)
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

	// 5. 下游微服务
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitClients(c.Services, clients)
	if err != nil {
		redisCache.Close()
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("grpc clients init error: %w", err)
	}

	// 6. DDD 分层装配
	bootLog.Info("assembling services with full dependency injection...")
	adminRepo := mysql.NewAdminRepository(db)
	roleRepo := mysql.NewRoleRepository(db)
	auditRepo := mysql.NewAuditRepository(db)
	approvalRepo := mysql.NewApprovalRepository(db)
	settingRepo := mysql.NewSettingRepository(db)

	opsDeps := application.SystemOpsDependencies{
		OrderClient:   clients.Order,
		UserClient:    clients.User,
		PaymentClient: clients.Payment,
		Storage:       store, // 【对齐】：Storage 像 DB/Redis 一样被注入
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

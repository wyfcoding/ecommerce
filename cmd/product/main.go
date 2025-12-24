package product

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/goapi/product/v1"
	"github.com/wyfcoding/ecommerce/internal/product/application"
	mysqlRepo "github.com/wyfcoding/ecommerce/internal/product/infrastructure/persistence/mysql"
	grpcServer "github.com/wyfcoding/ecommerce/internal/product/interfaces/grpc"
	producthttp "github.com/wyfcoding/ecommerce/internal/product/interfaces/http"
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
const BootstrapName = "product"

// AppContext 应用上下文，包含配置、服务实例和客户端依赖。
type AppContext struct {
	Config     *configpkg.Config
	AppService *application.ProductService
	Clients    *ServiceClients
}

// ServiceClients 包含所有下游服务的 gRPC 客户端连接。
type ServiceClients struct {
	Inventory      *grpc.ClientConn
	Review         *grpc.ClientConn
	Pricing        *grpc.ClientConn
	Recommendation *grpc.ClientConn
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
	pb.RegisterProductServer(s, grpcServer.NewServer(ctx.AppService))
}

func registerGin(e *gin.Engine, svc any) {
	ctx := svc.(*AppContext)
	handler := producthttp.NewHandler(ctx.AppService, slog.Default())
	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)
}

func initService(cfg any, m *metrics.Metrics) (any, func(), error) {
	c := cfg.(*configpkg.Config)
	slog.Info("initializing service dependencies...")

	// 1. 数据库
	db, err := databases.NewDB(c.Data.Database, logging.Default())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 2. Redis 缓存 & BigCache
	redisCache, err := cache.NewRedisCache(c.Data.Redis)
	if err != nil {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to connect redis: %w", err)
	}
	bigCache, err := cache.NewBigCache(c.Data.BigCache.LifeWindow, c.Data.BigCache.HardMaxCacheSize)
	if err != nil {
		redisCache.Close()
		sqlDB, _ := db.DB()
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init bigcache: %w", err)
	}
	multiLevelCache := cache.NewMultiLevelCache(bigCache, redisCache)

	// 3. 下游服务客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitServiceClients(c.Services, clients)
	if err != nil {
		multiLevelCache.Close()
		sqlDB, _ := db.DB()
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 4. 基础设施与应用层
	productRepo := mysqlRepo.NewProductRepository(db)
	skuRepo := mysqlRepo.NewSKURepository(db)
	categoryRepo := mysqlRepo.NewCategoryRepository(db)
	brandRepo := mysqlRepo.NewBrandRepository(db)

	logger := logging.Default().Logger
	catalogService := application.NewCatalogService(productRepo, multiLevelCache, logger, m)
	categoryService := application.NewCategoryService(categoryRepo, logger)
	brandService := application.NewBrandService(brandRepo, logger)
	skuService := application.NewSKUService(productRepo, skuRepo, logger)

	appService := application.NewProductService(
		catalogService,
		categoryService,
		brandService,
		skuService,
		logger,
	)

	cleanup := func() {
		slog.Info("cleaning up resources...")
		clientCleanup()
		multiLevelCache.Close()
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	return &AppContext{
		Config:     c,
		AppService: appService,
		Clients:    clients,
	}, cleanup, nil
}

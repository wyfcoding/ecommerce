package main

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/go-api/product/v1"
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

const BootstrapName = "product"

type AppContext struct {
	Config     *configpkg.Config
	AppService *application.ProductService
	Clients    *ServiceClients
}

type ServiceClients struct {
	Inventory      *grpc.ClientConn
	Review         *grpc.ClientConn
	Pricing        *grpc.ClientConn
	Recommendation *grpc.ClientConn
}

func main() {
	app.NewBuilder(BootstrapName).
		WithConfig(&configpkg.Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGinMiddleware(middleware.CORS()).
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, svc interface{}) {
	ctx := svc.(*AppContext)
	pb.RegisterProductServer(s, grpcServer.NewServer(ctx.AppService))
}

func registerGin(e *gin.Engine, svc interface{}) {
	ctx := svc.(*AppContext)
	handler := producthttp.NewHandler(ctx.AppService, slog.Default())
	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	c := cfg.(*configpkg.Config)
	slog.Info("initializing service dependencies...")

	// 1. Database
	db, err := databases.NewDB(c.Data.Database, logging.Default())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 2. Redis & BigCache
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

	// 3. Downstream Clients
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitServiceClients(c.Services, clients)
	if err != nil {
		multiLevelCache.Close()
		sqlDB, _ := db.DB()
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 4. Infrastructure & Application
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

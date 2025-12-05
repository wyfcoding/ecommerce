package main

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	v1 "github.com/wyfcoding/ecommerce/go-api/product/v1"
	"github.com/wyfcoding/ecommerce/internal/product/application"
	mysqlRepo "github.com/wyfcoding/ecommerce/internal/product/infrastructure/persistence/mysql"
	grpcServer "github.com/wyfcoding/ecommerce/internal/product/interfaces/grpc"
	producthttp "github.com/wyfcoding/ecommerce/internal/product/interfaces/http"
	"github.com/wyfcoding/ecommerce/pkg/app"
	"github.com/wyfcoding/ecommerce/pkg/cache"
	configpkg "github.com/wyfcoding/ecommerce/pkg/config"
	"github.com/wyfcoding/ecommerce/pkg/databases"
	"github.com/wyfcoding/ecommerce/pkg/logging"
	"github.com/wyfcoding/ecommerce/pkg/metrics"

	"google.golang.org/grpc"
)

type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

func main() {
	app.NewBuilder("product").
		WithConfig(&Config{}).
		WithService(func(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
			c := cfg.(*Config)

			// Database
			logger := logging.NewLogger("product", "app")
			db, err := databases.NewDB(c.Data.Database, logger)
			if err != nil {
				return nil, nil, err
			}

			sqlDB, err := db.DB()
			if err != nil {
				return nil, nil, err
			}

			cleanupDB := func() {
				sqlDB.Close()
			}

			// Repositories
			productRepo := mysqlRepo.NewProductRepository(db)
			skuRepo := mysqlRepo.NewSKURepository(db)
			categoryRepo := mysqlRepo.NewCategoryRepository(db)
			brandRepo := mysqlRepo.NewBrandRepository(db)

			// Caches
			redisCache, err := cache.NewRedisCache(c.Data.Redis)
			if err != nil {
				cleanupDB()
				return nil, nil, err
			}
			bigCache, err := cache.NewBigCache(c.Data.BigCache.LifeWindow, c.Data.BigCache.HardMaxCacheSize)
			if err != nil {
				cleanupDB()
				redisCache.Close()
				return nil, nil, err
			}
			multiLevelCache := cache.NewMultiLevelCache(bigCache, redisCache)

			cleanupCache := func() {
				multiLevelCache.Close()
			}

			// Application Layer
			appService := application.NewProductApplicationService(
				productRepo,
				skuRepo,
				categoryRepo,
				brandRepo,
				multiLevelCache,
				slog.Default(),
				m,
			)

			// gRPC Server
			srv := grpcServer.NewServer(appService)

			return srv, func() {
				cleanupCache()
				cleanupDB()
			}, nil
		}).
		WithGRPC(func(s *grpc.Server, svc interface{}) {
			v1.RegisterProductServer(s, svc.(v1.ProductServer))
		}).
		WithGin(registerGin).
		Build().
		Run()
}

func registerGin(e *gin.Engine, srv interface{}) {
	service := srv.(*application.ProductApplicationService)
	handler := producthttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered for product service (DDD)")
}

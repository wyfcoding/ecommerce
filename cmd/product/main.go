package main

import (
	v1 "ecommerce/api/product/v1"
	"ecommerce/internal/product/application"
	mysqlRepo "ecommerce/internal/product/infrastructure/persistence/mysql"
	grpcServer "ecommerce/internal/product/interfaces/grpc"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	"ecommerce/pkg/databases"
	"ecommerce/pkg/logging"
	"ecommerce/pkg/metrics"

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

			// Application Service
			appService := application.NewProductApplicationService(
				productRepo,
				skuRepo,
				categoryRepo,
				brandRepo,
			)

			// gRPC Server
			srv := grpcServer.NewServer(appService)

			return srv, func() {
				cleanupDB()
			}, nil
		}).
		WithGRPC(func(s *grpc.Server, svc interface{}) {
			v1.RegisterProductServer(s, svc.(v1.ProductServer))
		}).
		Build().
		Run()
}

package main

import (
	v1 "ecommerce/api/payment/v1"
	"ecommerce/internal/payment/application"
	mysqlRepo "ecommerce/internal/payment/infrastructure/persistence/mysql"
	grpcServer "ecommerce/internal/payment/interfaces/grpc"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	"ecommerce/pkg/databases"
	"ecommerce/pkg/idgen"
	"ecommerce/pkg/logging"
	"ecommerce/pkg/metrics"

	"google.golang.org/grpc"
)

type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

func main() {
	app.NewBuilder("payment").
		WithConfig(&Config{}).
		WithService(func(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
			c := cfg.(*Config)

			// Database
			logger := logging.NewLogger("payment-service", "app")
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
			paymentRepo := mysqlRepo.NewPaymentRepository(db)

			// ID Generator
			idGenerator, err := idgen.NewSnowflakeGenerator(c.Snowflake)
			if err != nil {
				cleanupDB()
				return nil, nil, err
			}

			// Application Service
			appService := application.NewPaymentApplicationService(
				paymentRepo,
				idGenerator,
			)

			// gRPC Server
			srv := grpcServer.NewServer(appService)

			return srv, func() {
				cleanupDB()
			}, nil
		}).
		WithGRPC(func(s *grpc.Server, svc interface{}) {
			v1.RegisterPaymentServer(s, svc.(v1.PaymentServer))
		}).
		Build().
		Run()
}

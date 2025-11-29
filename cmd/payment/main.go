package main

import (
	v1 "github.com/wyfcoding/ecommerce/api/payment/v1"
	"github.com/wyfcoding/ecommerce/internal/payment/application"
	"github.com/wyfcoding/ecommerce/internal/payment/infrastructure/persistence"
	grpcServer "github.com/wyfcoding/ecommerce/internal/payment/interfaces/grpc"
	"github.com/wyfcoding/ecommerce/pkg/app"
	configpkg "github.com/wyfcoding/ecommerce/pkg/config"
	"github.com/wyfcoding/ecommerce/pkg/databases/sharding"
	"github.com/wyfcoding/ecommerce/pkg/idgen"
	"github.com/wyfcoding/ecommerce/pkg/logging"
	"github.com/wyfcoding/ecommerce/pkg/metrics"

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

			// Database Sharding Manager
			logger := logging.NewLogger("payment-service", "app")
			shardingManager, err := sharding.NewManager(c.Data.Shards, logger)
			if err != nil {
				// Fallback to single DB
				if len(c.Data.Shards) == 0 {
					logger.Info("No shards configured, using single database connection")
					shardingManager, err = sharding.NewManager([]configpkg.DatabaseConfig{c.Data.Database}, logger)
					if err != nil {
						return nil, nil, err
					}
				} else {
					return nil, nil, err
				}
			}

			cleanupDB := func() {
				shardingManager.Close()
			}

			// Repositories
			paymentRepo := persistence.NewPaymentRepository(shardingManager)

			// ID Generator
			idGenerator, err := idgen.NewSnowflakeGenerator(c.Snowflake)
			if err != nil {
				cleanupDB()
				return nil, nil, err
			}

			// Application Service
			appService := application.NewPaymentApplicationService(paymentRepo, idGenerator, logger.Logger)

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

package main

import (
	pb "github.com/wyfcoding/ecommerce/api/user/v1"
	"github.com/wyfcoding/ecommerce/internal/user/application"
	mysqlRepo "github.com/wyfcoding/ecommerce/internal/user/infrastructure/persistence/mysql"
	usergrpc "github.com/wyfcoding/ecommerce/internal/user/interfaces/grpc"
	"github.com/wyfcoding/ecommerce/pkg/app"
	"github.com/wyfcoding/ecommerce/pkg/config"
	"github.com/wyfcoding/ecommerce/pkg/databases"
	"github.com/wyfcoding/ecommerce/pkg/logging"
	"github.com/wyfcoding/ecommerce/pkg/metrics"

	"google.golang.org/grpc"
)

type Config struct {
	config.Config `mapstructure:",squash"`
}

func main() {
	app.NewBuilder("user").
		WithConfig(&config.Config{}).
		WithService(func(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
			c := cfg.(*config.Config)

			// Initialize Logger
			logger := logging.NewLogger("user", "app")

			// Initialize DB
			db, err := databases.NewDB(c.Data.Database, logger)
			if err != nil {
				return nil, nil, err
			}
			dbCleanup := func() {
				if sqlDB, err := db.DB(); err == nil {
					sqlDB.Close()
				}
			}

			repo := mysqlRepo.NewUserRepository(db)
			// Address repo is also needed by UserApplicationService
			addressRepo := mysqlRepo.NewAddressRepository(db)

			svc := application.NewUserApplicationService(
				repo,
				addressRepo,
				c.JWT.Secret,
				c.JWT.Issuer,
				c.JWT.ExpireDuration,
				logger.Logger,
			)

			return svc, func() {
				dbCleanup()
			}, nil
		}).
		WithGRPC(func(s *grpc.Server, svc interface{}) {
			appSvc := svc.(*application.UserApplicationService)
			impl := usergrpc.NewServer(appSvc)
			pb.RegisterUserServer(s, impl)
		}).
		Build().
		Run()
}

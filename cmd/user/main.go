package main

import (
	pb "ecommerce/api/user/v1"
	"ecommerce/internal/user/application"
	mysqlRepo "ecommerce/internal/user/infrastructure/persistence/mysql"
	usergrpc "ecommerce/internal/user/interfaces/grpc"
	"ecommerce/pkg/app"
	"ecommerce/pkg/config"
	"ecommerce/pkg/databases"
	"ecommerce/pkg/logging"
	"ecommerce/pkg/metrics"

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
				c.JWT.Expire,
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

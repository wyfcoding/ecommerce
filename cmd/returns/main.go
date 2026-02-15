package main

import (
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/go-api/returns/v1"
	"github.com/wyfcoding/ecommerce/internal/returns/application"
	returnsmysql "github.com/wyfcoding/ecommerce/internal/returns/infrastructure/persistence/mysql"
	returnsgrpc "github.com/wyfcoding/ecommerce/internal/returns/interfaces/grpc"
	"github.com/wyfcoding/pkg/app"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"google.golang.org/grpc"
)

const BootstrapName = "returns"

type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

type AppContext struct {
	Config  *Config
	Service *application.ReturnService
}

func main() {
	if err := app.NewBuilder[*Config, *AppContext](BootstrapName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		Build().
		Run(); err != nil {
		slog.Error("returns service bootstrap failed", "error", err)
	}
}

func registerGRPC(s *grpc.Server, ctx *AppContext) {
	pb.RegisterReturnsServiceServer(s, returnsgrpc.NewServer(ctx.Service, slog.Default()))
}

func initService(cfg *Config, m *metrics.Metrics) (*AppContext, func(), error) {
	c := cfg
	bootLog := slog.With("module", "bootstrap")
	logger := logging.Default()

	configpkg.PrintWithMask(c)

	// 1. Database
	db, err := database.NewDB(c.Data.Database, c.CircuitBreaker, logger, m)
	if err != nil {
		return nil, nil, fmt.Errorf("database init error: %w", err)
	}

	// 2. Auto-migration (Dev)
	if c.Server.Environment == "dev" {
		bootLog.Info("running auto-migration for returns service")
		if err := db.RawDB().AutoMigrate(
			&returnsmysql.ReturnRequestModel{},
		); err != nil {
			bootLog.Error("failed to migrate database", "error", err)
		}
	}

	// 3. ID Generator
	idGen, err := idgen.NewGenerator(c.Snowflake)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init idgen: %w", err)
	}

	// 4. Repositories & Services
	returnRepo := returnsmysql.NewReturnRepository(db, logger)
	returnSvc := application.NewReturnService(returnRepo, idGen, logger.Logger)

	cleanup := func() {
		bootLog.Info("shutting down returns service...")
		if sqlDB, err := db.RawDB().DB(); err == nil && sqlDB != nil {
			sqlDB.Close()
		}
	}

	return &AppContext{
		Config:  c,
		Service: returnSvc,
	}, cleanup, nil
}

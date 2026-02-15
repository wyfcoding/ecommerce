package main

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	pb "github.com/wyfcoding/ecommerce/go-api/cms/v1"
	"github.com/wyfcoding/ecommerce/internal/cms/application"
	cmsmysql "github.com/wyfcoding/ecommerce/internal/cms/infrastructure/persistence/mysql"
	cmsgrpc "github.com/wyfcoding/ecommerce/internal/cms/interfaces/grpc"
	cmshttp "github.com/wyfcoding/ecommerce/internal/cms/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"google.golang.org/grpc"
)

const BootstrapName = "cms"

type Config struct {
	configpkg.Config `mapstructure:",squash"`
}

type AppContext struct {
	Config  *Config
	Service *application.CMSService
}

func main() {
	if err := app.NewBuilder[*Config, *AppContext](BootstrapName).
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerHTTP).
		Build().
		Run(); err != nil {
		slog.Error("cms service bootstrap failed", "error", err)
	}
}

func registerGRPC(s *grpc.Server, ctx *AppContext) {
	pb.RegisterCMSServiceServer(s, cmsgrpc.NewServer(ctx.Service, slog.Default()))
}

func registerHTTP(r *gin.Engine, ctx *AppContext) {
	cmshttp.NewHandler(ctx.Service, slog.Default()).RegisterRoutes(r)
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
		bootLog.Info("running auto-migration for cms service")
		if err := db.RawDB().AutoMigrate(
			&cmsmysql.PageModel{},
			&cmsmysql.TemplateModel{},
			&cmsmysql.AssetModel{},
		); err != nil {
			bootLog.Error("failed to migrate database", "error", err)
		}
	}

	// 3. ID Generator (Snowflake)
	idGen, err := idgen.NewGenerator(c.Snowflake)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init idgen: %w", err)
	}

	// 4. Repositories & Services
	cmsRepo := cmsmysql.NewCMSRepository(db, logger)
	cmsSvc := application.NewCMSService(cmsRepo, idGen, logger.Logger)

	cleanup := func() {
		bootLog.Info("shutting down cms service...")
		if sqlDB, err := db.RawDB().DB(); err == nil && sqlDB != nil {
			sqlDB.Close()
		}
	}

	return &AppContext{
		Config:  c,
		Service: cmsSvc,
	}, cleanup, nil
}

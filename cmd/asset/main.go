package main

import (
	"fmt"
	"time"

	"ecommerce/api/asset/v1"
	"ecommerce/internal/asset/biz"
	"ecommerce/internal/asset/data"
	"ecommerce/internal/asset/handler"
	"ecommerce/internal/asset/service"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	"ecommerce/pkg/database/mysql"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/minio"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
)

// Config is the service-specific configuration structure.
type Config struct {
	configpkg.Config
	Data struct {
		configpkg.DataConfig
		Database struct {
			configpkg.DatabaseConfig
			LogLevel      gormlogger.LogLevel `toml:"log_level"`
			SlowThreshold time.Duration     `toml:"slow_threshold"`
		} `toml:"database"`
		Minio minio.Config `toml:"minio"`
	} `toml:"data"`
}

func main() {
	app.NewBuilder("asset").
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithMetrics("9093").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	v1.RegisterAssetServiceServer(s, srv.(v1.AssetServiceServer))
}

func registerGin(e *gin.Engine, srv interface{}) {
	assetHandler := handler.NewAssetHandler(srv.(*service.AssetService))
	// e.g., e.POST("/v1/assets/upload", assetHandler.UploadAsset)
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)

	db, dbCleanup, err := mysql.NewGORMDB(&config.Data.Database, zap.L())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to mysql: %w", err)
	}

	minioClient, err := minio.NewMinioClient(&config.Data.Minio)
	if err != nil {
		dbCleanup()
		return nil, nil, fmt.Errorf("failed to connect to minio: %w", err)
	}

	dataInstance, dataCleanup := data.NewData(db)
	assetRepo := data.NewAssetRepo(dataInstance, minioClient)
	assetUsecase := biz.NewAssetUsecase(assetRepo)
	assetService := service.NewAssetService(assetUsecase)

	cleanup := func() {
		dataCleanup()
		dbCleanup()
	}

	return assetService, cleanup, nil
}
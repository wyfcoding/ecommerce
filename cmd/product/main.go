package main

import (
	"fmt"
	"time"

	"ecommerce/api/product/v1"
	"ecommerce/internal/product/biz"
	"ecommerce/internal/product/data"
	"ecommerce/internal/product/handler"
	"ecommerce/internal/product/service"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	"ecommerce/pkg/metrics"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
	} `toml:"data"`
}

func main() {
	app.NewBuilder("product").
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithMetrics("9091").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	v1.RegisterProductServer(s, srv.(v1.ProductServer))
}

func registerGin(e *gin.Engine, srv interface{}) {
	productHandler := handler.NewProductHandler(srv.(*service.ProductService))
	// Define routes here, e.g.:
	// e.GET("/v1/products/:id", productHandler.GetProduct)
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)

	db, err := gorm.Open(mysql.Open(config.Data.Database.DSN), &gorm.Config{
		Logger: gormlogger.New(zap.L(), config.Data.Database.LogLevel, config.Data.Database.SlowThreshold),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	if err := db.AutoMigrate(&data.Product{}, &data.Category{}); err != nil {
		return nil, nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	dataInstance := data.NewData(db)
	productRepo := data.NewProductRepo(dataInstance)
	categoryRepo := data.NewCategoryRepo(dataInstance)
	productUsecase := biz.NewProductUsecase(productRepo)
	categoryUsecase := biz.NewCategoryUsecase(categoryRepo)
	productService := service.NewProductService(productUsecase, categoryUsecase)

	cleanup := func() {
		sqlDB.Close()
	}

	return productService, cleanup, nil
}
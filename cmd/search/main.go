package main

import (
	"context"
	"fmt"
	"time"

	v1 "ecommerce/api-go/api/search/v1"
	"ecommerce/internal/search/biz"
	"ecommerce/internal/search/data"
	"ecommerce/internal/search/handler"
	"ecommerce/internal/search/service"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	"ecommerce/pkg/metrics"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
		ProductService struct {
			Addr string `toml:"addr"`
		}
	}
} `toml:"data"`

func main() {
	app.NewBuilder("search").
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithMetrics("9095").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	v1.RegisterSearchServiceServer(s, srv.(v1.SearchServiceServer))
}

func registerGin(e *gin.Engine, srv interface{}) {
	searchHandler := handler.NewSearchHandler(srv.(*service.SearchService))
	// e.g., e.GET("/v1/search", searchHandler.SearchProducts)
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)

	db, err := gorm.Open(mysql.Open(config.Data.Database.DSN), &gorm.Config{
		Logger: gormlogger.New(zap.L(), config.Data.Database.LogLevel, config.Data.Database.SlowThreshold),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	dialCtx, cancelDial := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelDial()
	productServiceConn, err := grpc.DialContext(dialCtx, config.Data.ProductService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to connect to product service: %w", err)
	}

	productClient := data.NewProductClient(productServiceConn)
	searchRepo := data.NewSearchRepo(db)
	searchUsecase := biz.NewSearchUsecase(searchRepo, productClient)
	searchService := service.NewSearchService(searchUsecase)

	cleanup := func() {
		sqlDB.Close()
		productServiceConn.Close()
	}

	return searchService, cleanup, nil
}

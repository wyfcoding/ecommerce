package main

import (
	"context"
	"fmt"
	"time"

	"ecommerce/api/search/v1"
	"ecommerce/internal/search/handler"
	"ecommerce/internal/search/repository"
	"ecommerce/internal/search/service"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	mysqlpkg "ecommerce/pkg/database/mysql"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/tracing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm/logger"
)

// Config is the service-specific configuration structure.
type Config struct {
	configpkg.Config
	Data struct {
		configpkg.DataConfig
		Database struct {
			configpkg.DatabaseConfig
			LogLevel      logger.LogLevel `toml:"log_level"`
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

	db, cleanupDB, err := mysqlpkg.NewGORMDB(&config.Data.Database)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 自动迁移数据库表结构 (TODO: 仅在开发环境或需要时执行)
	// if err := db.AutoMigate(&model.SearchQuery{}); err != nil {
	// 	return nil, nil, fmt.Errorf("failed to migrate database: %w", err)
	// }

	dialCtx, cancelDial := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelDial()
	productServiceConn, err := grpc.DialContext(dialCtx, config.Data.ProductService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		cleanupDB()
		return nil, nil, fmt.Errorf("failed to connect to product service: %w", err)
	}

	productClient := client.NewProductClient(productServiceConn)
	searchRepo := repository.NewSearchRepo(db)
	searchService := service.NewSearchService(searchRepo)

	cleanup := func() {
		cleanupDB()
		productServiceConn.Close()
	}

	return searchService, cleanup, nil
}

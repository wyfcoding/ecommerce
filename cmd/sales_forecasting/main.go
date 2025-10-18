package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	salesforecastinghandler "ecommerce/internal/sales_forecasting/handler"
	"ecommerce/internal/sales_forecasting/model"
	"ecommerce/internal/sales_forecasting/repository"
	"ecommerce/internal/sales_forecasting/service"
	configpkg "ecommerce/pkg/config"
	mysqlpkg "ecommerce/pkg/database/mysql"
	redisPkg "ecommerce/pkg/database/redis"
	"ecommerce/pkg/logging"
	"ecommerce/pkg/snowflake"
)

// Config 结构体用于映射 TOML 配置文件
type Config struct {
	configpkg.ServerConfig `toml:"server"`
	Data                   struct {
		Database mysqlpkg.Config `toml:"database"`
		Redis    redisPkg.Config `toml:"redis"`
	} `toml:"data"`
	Snowflake snowflake.Config `toml:"snowflake"`
	Log       logging.Config   `toml:"log"`
}

func main() {
	// 1. 加载配置
	var configPath string
	flag.StringVar(&configPath, "conf", "./configs/sales_forecasting.toml", "config file path")
	flag.Parse()

	var cfg Config
	if err := configpkg.LoadConfig(configPath, &cfg); err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// 2. 初始化日志
	logger := logging.NewLogger(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output)
	zap.ReplaceGlobals(logger)

	// 3. 初始化雪花算法
	snowflakeNode, err := snowflake.NewSnowflakeNode(&cfg.Snowflake)
	if err != nil {
		zap.S().Fatalf("failed to init snowflake: %v", err)
	}

	// 4. 依赖注入 (DI)
	// 数据库连接
	db, cleanupDB, err := mysqlpkg.NewGORMDB(&cfg.Data.Database)
	if err != nil {
		zap.S().Fatalf("failed to connect database: %v", err)
	}
	defer cleanupDB()

	// 自动迁移数据库表结构
	if err := db.AutoMigrate(&model.ForecastResult{}); err != nil {
		zap.S().Fatalf("failed to migrate database: %v", err)
	}

	// 初始化 Redis
	redisClient, redisCleanup, err := redisPkg.NewRedisClient(&cfg.Data.Redis)
	if err != nil {
		zap.S().Fatalf("failed to new redis client: %v", err)
	}
	defer redisCleanup()

	salesForecastingRepo := repository.NewSalesForecastingRepo(db, redisClient, snowflakeNode)
	salesForecastingService := service.NewSalesForecastingService(salesForecastingRepo)

	// 5. 启动 gRPC 和 HTTP Gateway
	grpcServer, grpcErrChan := salesforecastinghandler.StartGRPCServer(salesForecastingService, cfg.Server.GRPC.Addr, cfg.Server.GRPC.Port)
	if grpcServer == nil {
		zap.S().Fatalf("failed to start gRPC server: %v", <-grpcErrChan)
	}
	httpServer, httpErrChan := salesforecastinghandler.StartHTTPServer(context.Background(), cfg.Server.GRPC.Addr, cfg.Server.GRPC.Port, cfg.Server.HTTP.Addr, cfg.Server.HTTP.Port)
	if httpServer == nil {
		zap.S().Fatalf("failed to start HTTP server: %v", <-httpErrChan)
	}

	// 6. 等待中断信号或服务器错误以实现优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		zap.S().Info("Shutting down sales_forecasting service...")
	case err := <-grpcErrChan:
		zap.S().Errorf("gRPC server error: %v", err)
	case err := <-httpErrChan:
		zap.S().Errorf("HTTP server error: %v", err)
	}

	// 优雅地关闭服务器
	grpcServer.GracefulStop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zap.S().Errorf("HTTP server shutdown error: %v", err)
	}
}
